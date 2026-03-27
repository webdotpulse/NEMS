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

func (sc *StrategyController) Stop() {
	close(sc.stopCh)
}
