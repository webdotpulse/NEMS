package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
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
var strategyMapsMu sync.Mutex
var chargerCurrentSetpoints = make(map[int]float64)
var batteryForceChargeW = make(map[int]float64)
var evSmartChargeActive = make(map[int]bool)
var relayCurrentState = make(map[int]bool)
var batteryDischargeW = make(map[int]float64)
var inverterPowerLimitW = make(map[int]float64)

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

// InitStrategyController initializes the global StrategyCtrl instance.
func InitStrategyController() {
	StrategyCtrl = &StrategyController{
		stopCh: make(chan struct{}),
	}
}

// updatePricingCache fetches the current hour's EPEX spot price and determines
// if the current hour is among the 'N' cheapest hours of the day (for Smart EV logic).
// This reduces SQLite read operations during the 2-second control loop.
func (sc *StrategyController) updatePricingCache(settings models.SiteSettings) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	now := time.Now().In(loc)
	startOfHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc).UTC()

	var rawPrice float64
	err := db.QueryRow("SELECT price_per_kwh FROM epex_prices WHERE timestamp = ?", startOfHour).Scan(&rawPrice)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[ERROR] StrategyController: Error fetching current EPEX price: %v", err)
		rawPrice = 999.0
	} else if err == sql.ErrNoRows {
		rawPrice = 999.0
	}

	newPrice := CalculateEffectivePrice(now, rawPrice, settings)

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

// Start kicks off a background goroutine that polls the site settings
// and executes the core energy optimization loop every 2 seconds.
func (sc *StrategyController) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var settings models.SiteSettings
				row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, battery_grid_charge_strategy, force_charge_below_euro, force_discharge_above_euro, smart_ev_cheapest_hours, appliance_turn_on_excess_w, peak_shaving_buffer_w, peak_shaving_rampup_w, timezone, contract_type, fixed_price_peak_kwh, fixed_price_off_peak_kwh, fixed_inject_price_kwh, dynamic_markup_kwh, engie_markup_peak, engie_markup_off_peak, engie_markup_super_off_peak, engie_multiplier, engie_base_fee, custom_charge_schedule, superdal_optimization_enabled, superdal_target_soc FROM site_settings WHERE id = 1")
				err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.BatteryGridChargeStrategy, &settings.ForceChargeBelowEuro, &settings.ForceDischargeAboveEuro, &settings.SmartEvCheapestHours, &settings.ApplianceTurnOnExcessW, &settings.PeakShavingBufferW, &settings.PeakShavingRampupW, &settings.Timezone, &settings.ContractType, &settings.FixedPricePeakKwh, &settings.FixedPriceOffPeakKwh, &settings.FixedInjectPriceKwh, &settings.DynamicMarkupKwh, &settings.EngieMarkupPeak, &settings.EngieMarkupOffPeak, &settings.EngieMarkupSuperOffPeak, &settings.EngieMultiplier, &settings.EngieBaseFee, &settings.CustomChargeSchedule, &settings.SuperdalOptimizationEnabled, &settings.SuperdalTargetSoc)
				if err != nil {
					if err == sql.ErrNoRows {
						settings = models.SiteSettings{
							StrategyMode: "eco",
							CapacityPeakLimitKw: 2.5,
							ActiveInverterCurtailment: false,
							BatteryGridChargeStrategy: "price_only",
							ForceChargeBelowEuro: 0.0,
							ForceDischargeAboveEuro: 999.0,
							SmartEvCheapestHours: 0,
							ApplianceTurnOnExcessW: 0.0,
							PeakShavingBufferW: 200.0,
							PeakShavingRampupW: 500.0,
							Timezone: "Europe/Brussels",
							ContractType: "dynamic",
							FixedPricePeakKwh: 0.35,
							FixedPriceOffPeakKwh: 0.30,
							FixedInjectPriceKwh: 0.05,
							DynamicMarkupKwh: 0.15,
							EngieMarkupPeak: 0.15,
							EngieMarkupOffPeak: 0.15,
							EngieMarkupSuperOffPeak: 0.15,
							EngieMultiplier: 0.1448,
							EngieBaseFee: 0.0,
							CustomChargeSchedule: "[]",
							SuperdalOptimizationEnabled: false,
							SuperdalTargetSoc: 100.0,
						}
					} else {
						log.Printf("[ERROR] StrategyController: Error fetching site settings: %v", err)
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
				log.Println("[INFO] StrategyController stopped")
				return
			}
		}
	}()
}

// executeControlLoop calculates global site state based on current device metrics.
// It evaluates priority logic (Smart EV Charging, Arbitrage, Excess Solar Routing, Flanders Peak Shaving)
// and issues commands to controllable hardware (Chargers, Batteries, Inverters, Relays).
// Modifications to the `strategyMaps` are protected via `strategyMapsMu` to ensure thread safety.
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
		if data.Category == "meter" {
			if data.GridPowerW > 0 {
				totalGridImport += data.GridPowerW
			} else {
				totalGridExport += -data.GridPowerW
			}
		} else if data.Category == "inverter" {
			totalSolar += data.PowerW
			totalBattery += data.BatteryPowerW
			if data.HasGridMeter {
				if data.GridPowerW > 0 {
					totalGridImport += data.GridPowerW
				} else {
					totalGridExport += -data.GridPowerW
				}
			}
		} else if data.Category == "charger" {
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

	elapsedSec := now.Sub(startOfQuarter).Seconds()
	remSec := 900.0 - elapsedSec
	if remSec < 1.0 {
		remSec = 1.0 // prevent division by zero near the boundary
	}

	// Project the final 15-minute average assuming the current instantaneous power is maintained
	projectedPeak := (avg15MinImport*elapsedSec + totalGridImport*remSec) / 900.0
	SetProjectedQuarterPeakW(projectedPeak)

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
	tz := settings.Timezone
	if tz == "" {
		tz = "Europe/Brussels"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		loc = time.Local
	}
	currentTime := time.Now().In(loc)
	isSuperDal := currentTime.Hour() >= 1 && currentTime.Hour() < 7

	allowGridCharge := false
	if settings.BatteryGridChargeStrategy == "super_dal_only" {
		allowGridCharge = isSuperDal
	} else if settings.BatteryGridChargeStrategy == "hybrid" {
		allowGridCharge = isSuperDal || currentCachedPrice < settings.ForceChargeBelowEuro
	} else {
		// "price_only" or default
		allowGridCharge = currentCachedPrice < settings.ForceChargeBelowEuro
	}

	// Superdal Optimization Override
	superdalActive := false
	if settings.ContractType == "engie_flextime" && settings.SuperdalOptimizationEnabled && isSuperDal {
		allowGridCharge = true
		superdalActive = true
	}

	// Custom Charge Schedule Override
	type ScheduleSlot struct {
		Start     string  `json:"start"`
		End       string  `json:"end"`
		TargetSoc float64 `json:"target_soc"`
	}
	var schedule []ScheduleSlot
	customScheduleActive := false
	customScheduleTargetSoc := 100.0
	if settings.CustomChargeSchedule != "" && settings.CustomChargeSchedule != "[]" {
		err := json.Unmarshal([]byte(settings.CustomChargeSchedule), &schedule)
		if err == nil {
			currentMinutes := currentTime.Hour()*60 + currentTime.Minute()
			for _, slot := range schedule {
				var startH, startM, endH, endM int
				fmt.Sscanf(slot.Start, "%d:%d", &startH, &startM)
				fmt.Sscanf(slot.End, "%d:%d", &endH, &endM)
				startMin := startH*60 + startM
				endMin := endH*60 + endM

				if startMin < endMin {
					if currentMinutes >= startMin && currentMinutes < endMin {
						customScheduleActive = true
						customScheduleTargetSoc = slot.TargetSoc
						allowGridCharge = true
						break
					}
				} else { // crosses midnight
					if currentMinutes >= startMin || currentMinutes < endMin {
						customScheduleActive = true
						customScheduleTargetSoc = slot.TargetSoc
						allowGridCharge = true
						break
					}
				}
			}
		} else {
			log.Printf("[ERROR] Failed to parse custom charge schedule: %v", err)
		}
	}

	if allowGridCharge {
		for id, poller := range pollers {
			if _, ok := poller.(models.BatteryController); ok {
				dev := poller.GetDevice()
				if dev.HasBattery && dev.BatteryMode != "hold" {
					// Check SOC target blocks
					if customScheduleActive && cache[id].Soc >= customScheduleTargetSoc {
						continue // Target SOC reached for custom schedule
					}
					if !customScheduleActive && superdalActive && cache[id].Soc >= settings.SuperdalTargetSoc {
						continue // Target SOC reached for superdal
					}

					// Initially set intent to max grid capacity or physical limit
					// If we are in Flanders mode, available capacity is threshold - current import
					// If not Flanders, we can just command max available (100000)
					if settings.StrategyMode == "flanders" {
						buffer := settings.PeakShavingBufferW
						if buffer == 0 {
							buffer = 200 // Default fallback
						}
						threshold := (settings.CapacityPeakLimitKw * 1000) - buffer
						// Available instantaneous capacity is based on projecting the rest of the 15-minute window
						// targetRemPower = (threshold * 900 - avg15MinImport * elapsedSec) / remSec
						// we then subtract the current house load excluding the battery charge
						currentChargePower := cache[id].BatteryPowerW
						if currentChargePower < 0 {
							currentChargePower = 0 // Ignore discharging
						}
						targetRemPower := (threshold*900.0 - avg15MinImport*elapsedSec) / remSec
						availableGridCapacity := targetRemPower - (totalGridImport - currentChargePower)
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
			strategyMapsMu.Lock()
			currentSetpoint, exists := chargerCurrentSetpoints[id]
			strategyMapsMu.Unlock()
			if !exists {
				// Estimate current setpoint if unknown
				if cache[id].PowerW > 1000 {
					currentSetpoint = 16.0
				} else {
					currentSetpoint = 0.0
				}
				strategyMapsMu.Lock()
				chargerCurrentSetpoints[id] = currentSetpoint
				strategyMapsMu.Unlock()
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

		// Check if our projected load exceeds threshold
		// We use the projected 15-minute average to determine excess
		if projectedPeak > threshold {
			excess := (projectedPeak - threshold) * (900.0 / remSec)

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
		} else if projectedPeak < threshold - rampup {
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
					for id, poller := range pollers {
						if inverter, ok := poller.(models.InverterController); ok {
							strategyMapsMu.Lock()
							currentLimit, exists := inverterPowerLimitW[id]
							strategyMapsMu.Unlock()

							if !exists || math.Abs(currentLimit-totalHouseLoad) > 100.0 {
								log.Printf("[INFO] Netherlands Mode: Actively curtailing inverter. Setting limit to %.1f W", totalHouseLoad)
								err := inverter.SetActivePowerLimit(totalHouseLoad)
								if err == nil {
									strategyMapsMu.Lock()
									inverterPowerLimitW[id] = totalHouseLoad
									strategyMapsMu.Unlock()
								}
							}
						}
					}
				}
			} else if totalGridImport > 50 {
				// Release curtailment
				for id, poller := range pollers {
					if inverter, ok := poller.(models.InverterController); ok {
						strategyMapsMu.Lock()
						currentLimit, exists := inverterPowerLimitW[id]
						strategyMapsMu.Unlock()

						if !exists || currentLimit != 100000.0 {
							log.Printf("[INFO] Netherlands Mode: Releasing curtailment. Setting limit to 100000 W")
							err := inverter.SetActivePowerLimit(100000.0)
							if err == nil {
								strategyMapsMu.Lock()
								inverterPowerLimitW[id] = 100000.0
								strategyMapsMu.Unlock()
							}
						}
					}
				}
			}
		}
	}

	// Step 4: Issue Commands (Only if state changed to avoid Modbus spam)

	// Issue EV Charger Commands
	for id, desiredSp := range desiredEvSetpoints {
		if charger, ok := pollers[id].(models.ChargeController); ok {
			strategyMapsMu.Lock()
			currentSp := chargerCurrentSetpoints[id]
			strategyMapsMu.Unlock()
			if math.Abs(currentSp-desiredSp) > 0.5 || desiredSp == 0 && currentSp != 0 || desiredSp == 16 && currentSp != 16 {
				log.Printf("[INFO] Control Loop: Changing EV Charger %d from %.1f A to %.1f A", id, currentSp, desiredSp)
				err := charger.SetChargeCurrent(desiredSp)
				if err == nil {
					strategyMapsMu.Lock()
					chargerCurrentSetpoints[id] = desiredSp
					if desiredSp > 0 && isCachedCheapestHour {
						evSmartChargeActive[id] = true
					} else if desiredSp == 0 {
						evSmartChargeActive[id] = false
					}
					strategyMapsMu.Unlock()
				} else {
					log.Printf("[ERROR] Control Loop: Failed to execute EV charger command for %d: %v", id, err)
				}
			}
		}
	}

	// Issue Relay Commands
	for id, poller := range pollers {
		if relay, ok := poller.(models.RelayController); ok {
			desiredState := false
			if isCachedCheapestHour && settings.SmartEvCheapestHours > 0 {
				desiredState = true
			} else if settings.ApplianceTurnOnExcessW > 0 {
				if totalGridExport > settings.ApplianceTurnOnExcessW {
					desiredState = true
				}
			}

			strategyMapsMu.Lock()
			currentState, exists := relayCurrentState[id]
			strategyMapsMu.Unlock()

			if !exists || currentState != desiredState {
				log.Printf("[INFO] Control Loop: Changing Relay %d to %v", id, desiredState)
				err := relay.SetState(desiredState)
				if err == nil {
					strategyMapsMu.Lock()
					relayCurrentState[id] = desiredState
					strategyMapsMu.Unlock()
				} else {
					log.Printf("[ERROR] Control Loop: Failed to execute relay command for %d: %v", id, err)
				}
			}
		}
	}

	// Issue Battery Commands
	for id, poller := range pollers {
		if battery, ok := poller.(models.BatteryController); ok {
			dev := poller.GetDevice()
			if dev.HasBattery {
				desiredChargeW := desiredBatteryForceChargeW[id]
				strategyMapsMu.Lock()
				lastChargeW, exists := batteryForceChargeW[id]
				strategyMapsMu.Unlock()

				if !exists || (desiredChargeW == 0 && lastChargeW > 0) || (desiredChargeW > 0 && math.Abs(desiredChargeW-lastChargeW) > 50.0) {
					var err error
					if desiredChargeW > 0 {
						log.Printf("[INFO] Control Loop: Setting Battery Force Charge for %d to %.1f W", id, desiredChargeW)
						err = battery.ChargeBattery(desiredChargeW)
					} else if exists && lastChargeW > 0 {
						log.Printf("[INFO] Control Loop: Stopping Battery Force Charge for %d", id)
						err = battery.ChargeBattery(0)
					}

					if err == nil {
						strategyMapsMu.Lock()
						batteryForceChargeW[id] = desiredChargeW
						strategyMapsMu.Unlock()
					} else {
						log.Printf("[ERROR] Control Loop: Failed to execute battery charge command for %d: %v", id, err)
					}
				}

				// Handle discharge if needed (Flanders peak shave)
				dischargeW := desiredBatteryDischargeW[id]
				strategyMapsMu.Lock()
				lastDischargeW, dischExists := batteryDischargeW[id]
				strategyMapsMu.Unlock()

				if dischargeW > 0 {
					if !dischExists || math.Abs(dischargeW-lastDischargeW) > 50.0 {
						log.Printf("[INFO] Control Loop: Setting Battery Discharge for %d to %.1f W", id, dischargeW)
						err := battery.DischargeBattery(dischargeW)
						if err == nil {
							strategyMapsMu.Lock()
							batteryDischargeW[id] = dischargeW
							strategyMapsMu.Unlock()
						} else {
							log.Printf("[ERROR] Control Loop: Failed to execute battery discharge command for %d: %v", id, err)
						}
					}
				} else if dischExists && lastDischargeW > 0 {
					log.Printf("[INFO] Control Loop: Stopping Battery Discharge for %d", id)
					err := battery.DischargeBattery(0)
					if err == nil {
						strategyMapsMu.Lock()
						batteryDischargeW[id] = 0
						strategyMapsMu.Unlock()
					} else {
						log.Printf("[ERROR] Control Loop: Failed to execute battery stop discharge command for %d: %v", id, err)
					}
				}
			}
		}
	}
}

// applyNetherlandsMode evaluates zero-export constraints.
// It attempts to sink any excess solar back into home storage or EV chargers.
// If storage is full and EVs are not charging, it will curtail inverter production to exactly match house load.
func (sc *StrategyController) applyNetherlandsMode(activeCurtailment bool) {
	cache := PollerMgr.GetDeviceCache()
	pollers := PollerMgr.GetPollers()

	totalGrid := 0.0
	totalSolar := 0.0
	totalBattery := 0.0
	totalEvCharger := 0.0

	for _, data := range cache {
		if data.Category == "meter" {
			totalGrid += data.GridPowerW
		} else if data.Category == "inverter" {
			totalSolar += data.PowerW
			totalBattery += data.BatteryPowerW
			if data.HasGridMeter {
				totalGrid += data.GridPowerW
			}
		} else if data.Category == "charger" {
			totalEvCharger += data.PowerW
		}
	}

	totalGridExport := 0.0
	if totalGrid < 0 {
		totalGridExport = -totalGrid
	}

	if totalGridExport > 0 {
		log.Printf("[INFO] Netherlands Mode: Grid export (%.1f W). Attempting to sink power.", totalGridExport)

		// Action 1: Ramp up EV Chargers
		for id, poller := range pollers {
			if charger, ok := poller.(models.ChargeController); ok {
				strategyMapsMu.Lock()
				currentSetpoint, exists := chargerCurrentSetpoints[id]
				strategyMapsMu.Unlock()
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
						log.Printf("[INFO] Netherlands Mode: Ramping up charger %d current from %.1f A to %.1f A to sink export", id, currentSetpoint, newSetpoint)
						charger.SetChargeCurrent(newSetpoint)
						strategyMapsMu.Lock()
						chargerCurrentSetpoints[id] = newSetpoint
						strategyMapsMu.Unlock()

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
							log.Printf("[INFO] Netherlands Mode: Charging battery %d with %.1f W to sink export", id, totalGridExport)
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
				for id, poller := range pollers {
					if inverter, ok := poller.(models.InverterController); ok {
						strategyMapsMu.Lock()
						currentLimit, exists := inverterPowerLimitW[id]
						strategyMapsMu.Unlock()

						if !exists || math.Abs(currentLimit-totalHouseLoad) > 100.0 {
							log.Printf("[INFO] Netherlands Mode: Actively curtailing inverter. Setting limit to %.1f W", totalHouseLoad)
							err := inverter.SetActivePowerLimit(totalHouseLoad)
							if err == nil {
								strategyMapsMu.Lock()
								inverterPowerLimitW[id] = totalHouseLoad
								strategyMapsMu.Unlock()
							}
						}
					}
				}
			}
		} else if totalGrid > 50 {
			// Release curtailment. If totalGrid > 50, it means we are importing power from the grid,
			// which implies the house load has increased beyond our currently curtailed inverter limit.
			// Release the limit to cover the new load.
			for id, poller := range pollers {
				if inverter, ok := poller.(models.InverterController); ok {
					strategyMapsMu.Lock()
					currentLimit, exists := inverterPowerLimitW[id]
					strategyMapsMu.Unlock()

					if !exists || currentLimit != 100000.0 {
						log.Printf("[INFO] Netherlands Mode: Releasing curtailment. Setting limit to 100000 W")
						err := inverter.SetActivePowerLimit(100000.0)
						if err == nil {
							strategyMapsMu.Lock()
							inverterPowerLimitW[id] = 100000.0
							strategyMapsMu.Unlock()
						}
					}
				}
			}
		}
	}
}

// Stop safely shuts down the strategy evaluation background task.
func (sc *StrategyController) Stop() {
	close(sc.stopCh)
}
