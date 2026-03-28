package main

import (
	"database/sql"
	"log"
	"sync"
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
var batteryForceChargeW = make(map[int]float64)
var evSmartChargeActive = make(map[int]bool)

// Kwartierpiek (15-min sliding window) Tracking
type TimeValue struct {
	Time  time.Time
	Value float64
}
var gridImportSamples []TimeValue

var projectedQuarterPeakWMu sync.RWMutex
var projectedQuarterPeakW float64

func GetProjectedQuarterPeakW() float64 {
	projectedQuarterPeakWMu.RLock()
	defer projectedQuarterPeakWMu.RUnlock()
	return projectedQuarterPeakW
}

func SetProjectedQuarterPeakW(v float64) {
	projectedQuarterPeakWMu.Lock()
	defer projectedQuarterPeakWMu.Unlock()
	projectedQuarterPeakW = v
}

var currentCachedPrice float64 = 999.0
var isCachedCheapestHour bool = false
var lastPricingCacheUpdate time.Time
var lastSettingsForPricing models.SiteSettings

func InitStrategyController() {
	StrategyCtrl = &StrategyController{
		stopCh: make(chan struct{}),
	}
}

func (sc *StrategyController) updatePricingCache(settings models.SiteSettings) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	now := time.Now().In(loc)
	startOfHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc).UTC()

	var newPrice float64
	err := db.QueryRow("SELECT price_per_kwh FROM epex_prices WHERE timestamp = ?", startOfHour).Scan(&newPrice)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("StrategyController: Error fetching current EPEX price: %v", err)
		newPrice = 999.0
	} else if err == sql.ErrNoRows {
		newPrice = 999.0
	}

	newIsCheapest := false
	if settings.SmartEvCheapestHours > 0 {
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		endOfToday := startOfToday.AddDate(0, 0, 1)

		rows, err := db.Query("SELECT timestamp FROM epex_prices WHERE timestamp >= ? AND timestamp < ? ORDER BY price_per_kwh ASC LIMIT ?", startOfToday, endOfToday, settings.SmartEvCheapestHours)
		if err == nil {
			for rows.Next() {
				var ts time.Time
				if err := rows.Scan(&ts); err == nil {
					if ts.Equal(startOfHour) {
						newIsCheapest = true
						break
					}
				}
			}
			rows.Close()
		}
	}

	currentCachedPrice = newPrice
	isCachedCheapestHour = newIsCheapest
	lastPricingCacheUpdate = now
	lastSettingsForPricing = settings
}

func (sc *StrategyController) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var settings models.SiteSettings
				row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, force_charge_below_euro, smart_ev_cheapest_hours, appliance_turn_on_excess_w, peak_shaving_buffer_w, peak_shaving_rampup_w FROM site_settings WHERE id = 1")
				err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.ForceChargeBelowEuro, &settings.SmartEvCheapestHours, &settings.ApplianceTurnOnExcessW, &settings.PeakShavingBufferW, &settings.PeakShavingRampupW)
				if err != nil {
					if err == sql.ErrNoRows {
						settings = models.SiteSettings{
							StrategyMode: "eco",
							CapacityPeakLimitKw: 2.5,
							ActiveInverterCurtailment: false,
							ForceChargeBelowEuro: 0.0,
							SmartEvCheapestHours: 0,
							ApplianceTurnOnExcessW: 0.0,
							PeakShavingBufferW: 200.0,
							PeakShavingRampupW: 500.0,
						}
					} else {
						log.Printf("StrategyController: Error fetching site settings: %v", err)
						continue
					}
				}

				loc, _ := time.LoadLocation("Europe/Amsterdam")
				now := time.Now().In(loc)

				// Update pricing cache if the hour rolled over or settings changed
				if now.Hour() != lastPricingCacheUpdate.Hour() ||
					settings.SmartEvCheapestHours != lastSettingsForPricing.SmartEvCheapestHours {
					sc.updatePricingCache(settings)
				}

				sc.executeControlLoop(settings)

			case <-sc.stopCh:
				log.Println("StrategyController stopped")
				return
			}
		}
	}()
}

func (sc *StrategyController) executeControlLoop(settings models.SiteSettings) {
	cache := PollerMgr.GetDeviceCache()
	pollers := PollerMgr.GetPollers()

	// 1. Calculate base state
	var totalGridImport float64
	var totalGridExport float64
	var totalSolar float64
	var totalBattery float64
	var totalEvCharger float64

	for _, data := range cache {
		if data.Template == "homewizard_meter" || data.Template == "demo_dongle" {
			if data.GridPowerW > 0 {
				totalGridImport += data.GridPowerW
			} else {
				totalGridExport += -data.GridPowerW
			}
		} else if data.Template == "huawei_inverter" || data.Template == "solis_inverter" || data.Template == "sma_inverter" || data.Template == "demo_inverter" {
			totalSolar += data.PowerW
			totalBattery += data.BatteryPowerW
			if data.Template == "huawei_inverter" && data.HasGridMeter {
				if data.GridPowerW > 0 {
					totalGridImport += data.GridPowerW
				} else {
					totalGridExport += -data.GridPowerW
				}
			}
		} else if data.Template == "raedian_charger" || data.Template == "demo_charger" || data.Template == "alfen_charger" || data.Template == "easee_charger" || data.Template == "bender_charger" || data.Template == "peblar_charger" || data.Template == "phoenix_charger" {
			totalEvCharger += data.PowerW
		}
	}

	// Calculate desired setpoints instead of issuing commands immediately
	desiredEvSetpoints := make(map[int]float64)
	desiredBatteryForceChargeW := make(map[int]float64)
	desiredBatteryDischargeW := make(map[int]float64)

	// 1.5 Calculate Kwartierpiek (15-min synchronized average for Flanders mode)
	now := time.Now()
	quarterMin := (now.Minute() / 15) * 15
	startOfQuarter := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), quarterMin, 0, 0, now.Location())

	gridImportSamples = append(gridImportSamples, TimeValue{Time: now, Value: totalGridImport})
	var recentSamples []TimeValue
	var sumImport float64
	for _, sample := range gridImportSamples {
		if !sample.Time.Before(startOfQuarter) {
			recentSamples = append(recentSamples, sample)
			sumImport += sample.Value
		}
	}
	gridImportSamples = recentSamples
	avg15MinImport := 0.0
	if len(gridImportSamples) > 0 {
		avg15MinImport = sumImport / float64(len(gridImportSamples))
	}
	SetProjectedQuarterPeakW(avg15MinImport)

	// Step 2: Apply Dynamic Tariffs Base Intent

	// Battery Arbitrage Discharge Intent
	if currentCachedPrice > settings.ForceDischargeAboveEuro {
		for id, poller := range pollers {
			if _, ok := poller.(models.BatteryController); ok {
				dev := poller.GetDevice()
				if dev.HasBattery && dev.BatteryMode != "hold" && dev.BatteryMode != "force_charge" {
					desiredBatteryDischargeW[id] = 100000.0
				}
			}
		}
	}

	// Battery Arbitrage Intent
	if currentCachedPrice < settings.ForceChargeBelowEuro {
		for id, poller := range pollers {
			if _, ok := poller.(models.BatteryController); ok {
				dev := poller.GetDevice()
				if dev.HasBattery && dev.BatteryMode != "hold" {
					// Initially set intent to max grid capacity or physical limit
					// If we are in Flanders mode, available capacity is threshold - current import
					// If not Flanders, we can just command max available (100000)
					if settings.StrategyMode == "flanders" {
						buffer := settings.PeakShavingBufferW
						if buffer == 0 {
							buffer = 200 // Default fallback
						}
						threshold := (settings.CapacityPeakLimitKw * 1000) - buffer
						// Add current battery charge power back to available capacity calculation
						// because if it is currently charging, that power is already included in avg15MinImport
						currentChargePower := cache[id].BatteryPowerW
						if currentChargePower < 0 {
							currentChargePower = 0 // Ignore discharging
						}
						availableGridCapacity := threshold - avg15MinImport + currentChargePower
						if availableGridCapacity < 0 {
							availableGridCapacity = 0
						}
						desiredBatteryForceChargeW[id] = availableGridCapacity
					} else {
						desiredBatteryForceChargeW[id] = 100000.0
					}
				}
			}
		}
	}

	// Smart EV Charging Intent
	for id, poller := range pollers {
		if _, ok := poller.(models.ChargeController); ok {
			dev := poller.GetDevice()
			currentSetpoint, exists := chargerCurrentSetpoints[id]
			if !exists {
				// Estimate current setpoint if unknown
				if cache[id].PowerW > 1000 {
					currentSetpoint = 16.0
				} else {
					currentSetpoint = 0.0
				}
				chargerCurrentSetpoints[id] = currentSetpoint
			}

			// Start with current setpoint
			desiredEvSetpoints[id] = currentSetpoint

			chargeMode := dev.ChargeMode
			if chargeMode == "" {
				chargeMode = "eco"
			}

			if chargeMode == "off" {
				desiredEvSetpoints[id] = 0.0
			} else if chargeMode == "now" {
				desiredEvSetpoints[id] = 16.0
			} else if chargeMode == "pv_only" || chargeMode == "eco" {
				// Default behavior for Smart EV
				if chargeMode == "eco" && settings.SmartEvCheapestHours > 0 {
					if isCachedCheapestHour {
						// If cheapest hour, gently ramp up to max
						if desiredEvSetpoints[id] < 16.0 {
							newSetpoint := desiredEvSetpoints[id] + 1.0
							if newSetpoint < 6.0 && desiredEvSetpoints[id] == 0 {
								newSetpoint = 6.0
							}
							if newSetpoint > 16.0 {
								newSetpoint = 16.0
							}
							desiredEvSetpoints[id] = newSetpoint
						}
					} else {
						// Outside cheapest hours, block grid charging by aggressively dropping to 0
						// unless overridden by solar routing later
						desiredEvSetpoints[id] = 0.0
					}
				} else if chargeMode == "pv_only" {
					// PV only blocks all non-solar charging
					desiredEvSetpoints[id] = 0.0
				}
			}
		}
	}

	// Step 3: Global Solar Excess Routing
	// If there is excess solar, sink it into EVs regardless of the strategy mode.
	if totalGridExport > 0 {
		for id := range desiredEvSetpoints {
			dev := pollers[id].GetDevice()
			chargeMode := dev.ChargeMode
			if chargeMode == "off" || chargeMode == "now" {
				continue // Skip routing for Off/Now
			}

			if desiredEvSetpoints[id] < 16.0 {
				ampsToAdd := totalGridExport / 230.0
				newSetpoint := desiredEvSetpoints[id] + ampsToAdd
				if newSetpoint > 16.0 {
					newSetpoint = 16.0
				}
				if newSetpoint > 0 && newSetpoint < 6.0 {
					newSetpoint = 6.0
				}
				if newSetpoint > desiredEvSetpoints[id] {
					reductionW := (newSetpoint - desiredEvSetpoints[id]) * 230.0
					desiredEvSetpoints[id] = newSetpoint
					totalGridExport -= reductionW
				}
			}
			if totalGridExport <= 0 {
				break
			}
		}
	}

	// Step 4: Apply Strategy Overrides (Flanders / Netherlands)

	if settings.StrategyMode == "flanders" {
		buffer := settings.PeakShavingBufferW
		if buffer == 0 {
			buffer = 200
		}
		threshold := (settings.CapacityPeakLimitKw * 1000) - buffer
		rampup := settings.PeakShavingRampupW
		if rampup == 0 {
			rampup = 500
		}

		// Check if our current (or intended) load exceeds threshold
		// We use current grid import to determine excess
		if totalGridImport > threshold {
			excess := totalGridImport - threshold

			// First priority: reduce EV chargers
			for id := range desiredEvSetpoints {
				if desiredEvSetpoints[id] > 0 {
					ampsToReduce := excess / 230.0
					if ampsToReduce < 1 {
						ampsToReduce = 1
					}
					newSetpoint := desiredEvSetpoints[id] - ampsToReduce
					if newSetpoint < 6 {
						newSetpoint = 0
					}
					reductionW := (desiredEvSetpoints[id] - newSetpoint) * 230.0
					desiredEvSetpoints[id] = newSetpoint
					excess -= reductionW
					if excess <= 0 {
						break
					}
				}
			}

			// Second priority: throttle battery force charging if still excess
			if excess > 0 {
				for id, chargeW := range desiredBatteryForceChargeW {
					if chargeW > 0 {
						if chargeW > excess {
							desiredBatteryForceChargeW[id] -= excess
							excess = 0
							break
						} else {
							excess -= chargeW
							desiredBatteryForceChargeW[id] = 0
						}
					}
				}
			}

			// Third priority: discharge battery to cover house load
			if excess > 0 {
				for id, poller := range pollers {
					if _, ok := poller.(models.BatteryController); ok {
						dev := poller.GetDevice()
						if desiredBatteryForceChargeW[id] == 0 && dev.BatteryMode != "hold" && dev.BatteryMode != "force_charge" {
							desiredBatteryDischargeW[id] = excess
							excess = 0
							break
						}
					}
				}
			}
		} else if totalGridImport < threshold - rampup {
			// When house load drops, slowly ramp EV chargers UP if they were throttled
			// (If they are 0 because of SmartEV blocking, do not ramp up)
			for id, currentSp := range desiredEvSetpoints {
				// If Smart EV is blocking grid charge, don't auto-ramp in Flanders unless it's a cheapest hour
				if settings.SmartEvCheapestHours > 0 && !isCachedCheapestHour {
					continue
				}

				if currentSp < 16.0 {
					newSetpoint := currentSp + 1.0
					if newSetpoint < 6 && currentSp == 0 {
						newSetpoint = 6
					}
					if newSetpoint > 16 {
						newSetpoint = 16
					}
					desiredEvSetpoints[id] = newSetpoint
				}
			}
		}
	} else if settings.StrategyMode == "netherlands" {
		// Active Curtailment
		if settings.ActiveInverterCurtailment {
			if totalGridExport > 50 {
				batteriesFull := true
				for id := range pollers {
					if _, ok := pollers[id].(models.BatteryController); ok {
						if cache[id].Soc < 100 {
							batteriesFull = false
							break
						}
					}
				}
				if batteriesFull && totalEvCharger < 100 {
					totalHouseLoad := totalSolar + totalBattery + (totalGridImport - totalGridExport)
					for _, poller := range pollers {
						if inverter, ok := poller.(models.InverterController); ok {
							inverter.SetActivePowerLimit(totalHouseLoad)
						}
					}
				}
			} else if totalGridImport > 50 {
				// Release curtailment
				for _, poller := range pollers {
					if inverter, ok := poller.(models.InverterController); ok {
						inverter.SetActivePowerLimit(100000)
					}
				}
			}
		}
	}

	// Step 4: Issue Commands (Only if state changed to avoid Modbus spam)

	// Issue EV Charger Commands
	for id, desiredSp := range desiredEvSetpoints {
		if charger, ok := pollers[id].(models.ChargeController); ok {
			currentSp := chargerCurrentSetpoints[id]
			if currentSp != desiredSp {
				log.Printf("Control Loop: Changing EV Charger %d from %.1f A to %.1f A", id, currentSp, desiredSp)
				err := charger.SetChargeCurrent(desiredSp)
				if err == nil {
					chargerCurrentSetpoints[id] = desiredSp
					if desiredSp > 0 && isCachedCheapestHour {
						evSmartChargeActive[id] = true
					} else if desiredSp == 0 {
						evSmartChargeActive[id] = false
					}
				}
			}
		}
	}

	// Issue Relay Commands
	for _, poller := range pollers {
		if relay, ok := poller.(models.RelayController); ok {
			if isCachedCheapestHour && settings.SmartEvCheapestHours > 0 {
				relay.SetState(true)
			} else if settings.ApplianceTurnOnExcessW > 0 {
				if totalGridExport > settings.ApplianceTurnOnExcessW {
					relay.SetState(true)
				} else {
					relay.SetState(false)
				}
			} else {
				relay.SetState(false)
			}
		}
	}

	// Issue Battery Commands
	for id, poller := range pollers {
		if battery, ok := poller.(models.BatteryController); ok {
			dev := poller.GetDevice()
			if dev.HasBattery {
				desiredChargeW := desiredBatteryForceChargeW[id]
				lastChargeW, exists := batteryForceChargeW[id]

				if !exists || desiredChargeW != lastChargeW {
					if desiredChargeW > 0 {
						log.Printf("Control Loop: Setting Battery Force Charge for %d to %.1f W", id, desiredChargeW)
						battery.ChargeBattery(desiredChargeW)
					} else if exists && lastChargeW > 0 {
						log.Printf("Control Loop: Stopping Battery Force Charge for %d", id)
						battery.ChargeBattery(0)
					}
					batteryForceChargeW[id] = desiredChargeW
				}

				// Handle discharge if needed (Flanders peak shave)
				dischargeW := desiredBatteryDischargeW[id]
				if dischargeW > 0 {
					// For simplicity we just issue discharge command if requested,
					// assuming the poller handles throttling or repeated commands fine,
					// or we could track last discharge command.
					battery.DischargeBattery(dischargeW)
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
