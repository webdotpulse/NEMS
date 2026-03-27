package main

import (
	"database/sql"
	"log"
	"time"

	"nems/internal/models"
)

type StrategyController struct {
	stopCh chan struct{}
}

var StrategyCtrl *StrategyController

// Assuming a global map to track charger setpoints for simulation purposes.
// In a real application, the charger poller should return its current setpoint,
// or it should be persisted in a database or device state.
var chargerCurrentSetpoints = make(map[int]float64)

func InitStrategyController() {
	StrategyCtrl = &StrategyController{
		stopCh: make(chan struct{}),
	}
}

func (sc *StrategyController) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var settings models.SiteSettings
				row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment FROM site_settings WHERE id = 1")
				err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment)
				if err != nil {
					if err == sql.ErrNoRows {
						settings = models.SiteSettings{
							StrategyMode: "eco",
							CapacityPeakLimitKw: 2.5,
							ActiveInverterCurtailment: false,
						}
					} else {
						log.Printf("StrategyController: Error fetching site settings: %v", err)
						continue
					}
				}
				if settings.StrategyMode == "flanders" {
					sc.applyFlandersMode(settings.CapacityPeakLimitKw)
				} else if settings.StrategyMode == "netherlands" {
					sc.applyNetherlandsMode(settings.ActiveInverterCurtailment)
				}
			case <-sc.stopCh:
				log.Println("StrategyController stopped")
				return
			}
		}
	}()
}

func (sc *StrategyController) applyFlandersMode(capacityPeakLimitKw float64) {
	threshold := (capacityPeakLimitKw * 1000) - 200
	cache := PollerMgr.GetDeviceCache()
	pollers := PollerMgr.GetPollers()

	var totalGridImport float64
	for _, data := range cache {
		if data.Template == "homewizard_meter" || data.Template == "demo_dongle" {
			// Assuming positive value is import, negative is export based on typical conventions or context.
			// If context implies GridPowerW is positive for import, we sum it.
			if data.GridPowerW > 0 {
				totalGridImport += data.GridPowerW
			}
		}
	}

	if totalGridImport > threshold {
		log.Printf("Flanders Mode: Grid import (%.1f W) exceeds threshold (%.1f W). Taking action.", totalGridImport, threshold)

		excess := totalGridImport - threshold
		chargerReduced := false

		// Action 1: Reduce EV Chargers
		for id, poller := range pollers {
			data := cache[id]
			if charger, ok := poller.(models.ChargeController); ok {
				currentSetpoint, exists := chargerCurrentSetpoints[id]
				if !exists {
					// Assume it's charging at 16A initially if it's consuming significant power
					if data.PowerW > 1000 {
						currentSetpoint = 16.0
					} else {
						currentSetpoint = 0.0
					}
				}

				if currentSetpoint > 0 {
					// Calculate how much we need to reduce. Rough estimate: 230V * 3 phases = 690W per Amp.
					// Let's use 230W per Amp (assuming 1-phase for safety/finer granularity) or just a fixed step.
					ampsToReduce := excess / 230.0
					if ampsToReduce < 1 {
						ampsToReduce = 1 // At least reduce by 1A
					}

					newSetpoint := currentSetpoint - ampsToReduce
					if newSetpoint < 6 {
						// Minimum charge current is usually 6A. If we must go below, pause it.
						newSetpoint = 0
					}

					log.Printf("Flanders Mode: Reducing charger %d current from %.1f A to %.1f A", id, currentSetpoint, newSetpoint)
					charger.SetChargeCurrent(newSetpoint)
					chargerCurrentSetpoints[id] = newSetpoint
					chargerReduced = true

					// Estimate power reduction (very roughly)
					reductionW := (currentSetpoint - newSetpoint) * 230.0
					excess -= reductionW

					if excess <= 0 {
						break
					}
				}
			}
		}

		// Action 2: Discharge Battery
		if excess > 0 && (!chargerReduced || excess > 0) {
			for id, poller := range pollers {
				if battery, ok := poller.(models.BatteryController); ok {
					dev := poller.GetDevice()
					if dev.HasBattery {
						// Let's poll for SOC.
						_, _, _, _, soc, err := poller.Poll()
						if err == nil && soc > 10 {
							log.Printf("Flanders Mode: Discharging battery %d to cover %.1f W", id, excess)
							battery.DischargeBattery(excess)
							excess = 0
							break
						}
					}
				}
			}
		}
	} else if totalGridImport < threshold - 500 {
		// When house load drops, slowly ramp the EV charger back up
		for id, poller := range pollers {
			if charger, ok := poller.(models.ChargeController); ok {
				currentSetpoint, exists := chargerCurrentSetpoints[id]
				if !exists {
					currentSetpoint = 0.0
				}

				// Maximum charge current is usually 16A
				if currentSetpoint < 16.0 {
					// Ramp up slowly by 1A every few cycles (since this runs every 2s, we don't want to spike it instantly)
					// Let's increment by 1A per tick to slowly increase load.
					newSetpoint := currentSetpoint + 1.0
					if newSetpoint < 6 && currentSetpoint == 0 {
						newSetpoint = 6 // Jump to min start current if it was paused
					}
					if newSetpoint > 16 {
						newSetpoint = 16
					}

					log.Printf("Flanders Mode: Ramping up charger %d current from %.1f A to %.1f A", id, currentSetpoint, newSetpoint)
					charger.SetChargeCurrent(newSetpoint)
					chargerCurrentSetpoints[id] = newSetpoint
				}
			}
		}
	}
}

func (sc *StrategyController) applyNetherlandsMode(activeCurtailment bool) {
	cache := PollerMgr.GetDeviceCache()
	pollers := PollerMgr.GetPollers()

	totalGrid := 0.0
	totalSolar := 0.0
	totalBattery := 0.0
	totalEvCharger := 0.0

	for _, data := range cache {
		if data.Template == "homewizard_meter" || data.Template == "demo_dongle" {
			totalGrid += data.GridPowerW
		} else if data.Template == "huawei_inverter" || data.Template == "solis_inverter" || data.Template == "sma_inverter" || data.Template == "demo_inverter" {
			totalSolar += data.PowerW
			totalBattery += data.BatteryPowerW
			if data.Template == "huawei_inverter" && data.HasGridMeter {
				totalGrid += data.GridPowerW
			}
		} else if data.Template == "raedian_charger" || data.Template == "demo_charger" || data.Template == "alfen_charger" || data.Template == "easee_charger" || data.Template == "bender_charger" || data.Template == "peblar_charger" || data.Template == "phoenix_charger" {
			totalEvCharger += data.PowerW
		}
	}

	totalGridExport := 0.0
	if totalGrid < 0 {
		totalGridExport = -totalGrid
	}

	if totalGridExport > 0 {
		log.Printf("Netherlands Mode: Grid export (%.1f W). Attempting to sink power.", totalGridExport)

		// Action 1: Ramp up EV Chargers
		for id, poller := range pollers {
			if charger, ok := poller.(models.ChargeController); ok {
				currentSetpoint, exists := chargerCurrentSetpoints[id]
				if !exists {
					currentSetpoint = 0.0
				}

				if currentSetpoint < 16.0 {
					ampsToAdd := totalGridExport / 230.0
					newSetpoint := currentSetpoint + ampsToAdd
					if newSetpoint > 16.0 {
						newSetpoint = 16.0
					}
					if newSetpoint > 0 && newSetpoint < 6.0 {
						newSetpoint = 6.0
					}

					if newSetpoint > currentSetpoint {
						log.Printf("Netherlands Mode: Ramping up charger %d current from %.1f A to %.1f A to sink export", id, currentSetpoint, newSetpoint)
						charger.SetChargeCurrent(newSetpoint)
						chargerCurrentSetpoints[id] = newSetpoint

						reductionW := (newSetpoint - currentSetpoint) * 230.0
						totalGridExport -= reductionW
					}
				}

				if totalGridExport <= 0 {
					break
				}
			}
		}

		// Action 2: Charge Home Battery
		if totalGridExport > 0 {
			for id, poller := range pollers {
				if battery, ok := poller.(models.BatteryController); ok {
					dev := poller.GetDevice()
					if dev.HasBattery {
						soc := cache[id].Soc
						if soc < 100 {
							log.Printf("Netherlands Mode: Charging battery %d with %.1f W to sink export", id, totalGridExport)
							battery.ChargeBattery(totalGridExport)
							totalGridExport = 0
							break
						}
					}
				}
			}
		}
	}

	// Action 3: Active Curtailment
	if activeCurtailment {
		if totalGridExport > 50 {
			// Check conditions for curtailment: all batteries full and EVs not charging
			batteriesFull := true
			for id, poller := range pollers {
				if _, ok := poller.(models.BatteryController); ok {
					dev := poller.GetDevice()
					if dev.HasBattery {
						soc := cache[id].Soc
						if soc < 100 {
							batteriesFull = false
							break
						}
					}
				}
			}

			if batteriesFull && totalEvCharger < 100 {
				totalHouseLoad := totalSolar + totalBattery + totalGrid
				for _, poller := range pollers {
					if inverter, ok := poller.(models.InverterController); ok {
						log.Printf("Netherlands Mode: Actively curtailing inverter. Setting limit to %.1f W", totalHouseLoad)
						inverter.SetActivePowerLimit(totalHouseLoad)
					}
				}
			}
		} else if totalGrid > 50 {
			// Release curtailment. If totalGrid > 50, it means we are importing power from the grid,
			// which implies the house load has increased beyond our currently curtailed inverter limit.
			// Release the limit to cover the new load.
			for _, poller := range pollers {
				if inverter, ok := poller.(models.InverterController); ok {
					inverter.SetActivePowerLimit(100000)
				}
			}
		}
	}
}

func (sc *StrategyController) Stop() {
	close(sc.stopCh)
}
