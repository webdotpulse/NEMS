package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"nems/internal/models"
	"nems/internal/templates"
)

// ---------------------------------------------------------
// Poller Manager
// ---------------------------------------------------------


type DeviceData struct {
	PowerW        float64
	BatteryPowerW float64
	GridPowerW    float64
	EnergyKwh     float64
	Soc           float64
	Status        string
	Template      string
	HasGridMeter  bool
	HasBattery    bool
}

type BufferedMeasurement struct {
	DeviceID  int
	PowerW    float64
	EnergyKwh float64
}

type PollerManager struct {
	pollers map[int]models.DevicePoller
	mu      sync.Mutex
	stopCh  chan struct{}

	bufferMu sync.Mutex
	buffer   []BufferedMeasurement

	cacheMu     sync.Mutex
	deviceCache map[int]DeviceData
}

var PollerMgr *PollerManager

func (pm *PollerManager) GetDeviceCache() map[int]DeviceData {
	pm.cacheMu.Lock()
	defer pm.cacheMu.Unlock()
	copyCache := make(map[int]DeviceData)
	for k, v := range pm.deviceCache {
		copyCache[k] = v
	}
	return copyCache
}

func (pm *PollerManager) GetPollers() map[int]models.DevicePoller {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	copyPollers := make(map[int]models.DevicePoller)
	for k, v := range pm.pollers {
		copyPollers[k] = v
	}
	return copyPollers
}

func InitPollerManager() {
	PollerMgr = &PollerManager{
		pollers: make(map[int]models.DevicePoller),
		stopCh:  make(chan struct{}),
		buffer:  make([]BufferedMeasurement, 0),
		deviceCache: make(map[int]DeviceData),
	}
}

func (pm *PollerManager) SyncDevices() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	log.Println("PollerManager: Syncing devices...")

	// Fetch current devices from DB
	rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity FROM devices")
	if err != nil {
		log.Printf("PollerManager: Error fetching devices: %v", err)
		return
	}
	defer rows.Close()

	activeDeviceIDs := make(map[int]bool)

	for rows.Next() {
		var d models.Device
		var username sql.NullString
		var password sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password, &d.HasGridMeter, &d.HasBattery, &d.BatteryCapacity); err != nil {
			log.Printf("PollerManager: Error scanning device: %v", err)
			continue
		}
		if username.Valid {
			d.Username = username.String
		}
		if password.Valid {
			d.Password = password.String
		}
		activeDeviceIDs[d.ID] = true

		// If poller doesn't exist, create it
		if _, exists := pm.pollers[d.ID]; !exists {
			poller := templates.CreatePoller(d.Template, d)
			if poller == nil {
				log.Printf("PollerManager: Unknown template %s for device %d", d.Template, d.ID)
				continue
			}

			err := poller.Connect()
			if err != nil {
				log.Printf("PollerManager: Failed to connect device %d: %v", d.ID, err)
			}
			pm.pollers[d.ID] = poller
			log.Printf("PollerManager: Added poller for device %d (%s)", d.ID, d.Name)
		}
	}

	// Remove old pollers
	for id, poller := range pm.pollers {
		if !activeDeviceIDs[id] {
			poller.Close()
			delete(pm.pollers, id)

			pm.cacheMu.Lock()
			delete(pm.deviceCache, id)
			pm.cacheMu.Unlock()
			log.Printf("PollerManager: Removed poller for device %d", id)
		}
	}
}

func (pm *PollerManager) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		fastTicker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		defer fastTicker.Stop()

		for {
			select {
			case <-fastTicker.C:
				pm.mu.Lock()
				polledAny := false

				for id, poller := range pm.pollers {
					device := poller.GetDevice()
					if device.Template != "homewizard_meter" {
						continue
					}
					polledAny = true

					powerW, batteryPowerW, gridPowerW, energyKwh, soc, err := poller.Poll()
					if err != nil {
						log.Printf("PollerManager: Error polling device %d: %v", id, err)

						pm.cacheMu.Lock()
						pm.deviceCache[id] = DeviceData{
							PowerW:        0,
							BatteryPowerW: 0,
							GridPowerW:    0,
							EnergyKwh:     0, // Note: energy is cumulative, but for live display, power is what matters. Energy is only used for history averaging, which we handle by not buffering on error.
							Soc:           0,
							Status:        poller.Status(),
							Template:      device.Template,
							HasGridMeter:  device.HasGridMeter,
							HasBattery:    device.HasBattery,
						}
						pm.cacheMu.Unlock()
						continue
					}

					pm.cacheMu.Lock()
					pm.deviceCache[id] = DeviceData{
						PowerW:        powerW,
						BatteryPowerW: batteryPowerW,
						GridPowerW:    gridPowerW,
						EnergyKwh:     energyKwh,
						Soc:           soc,
						Status:        poller.Status(),
						Template:      device.Template,
						HasGridMeter:  device.HasGridMeter,
						HasBattery:    device.HasBattery,
					}
					pm.cacheMu.Unlock()

					// Buffer measurement
					pm.bufferMu.Lock()
					pm.buffer = append(pm.buffer, BufferedMeasurement{
						DeviceID:  id,
						PowerW:    powerW,
						EnergyKwh: energyKwh,
					})
					pm.bufferMu.Unlock()
				}
				pm.mu.Unlock()

				if polledAny {
					pm.broadcastState()
				}

			case <-ticker.C:
				pm.mu.Lock()

				for id, poller := range pm.pollers {
					device := poller.GetDevice()
					if device.Template == "homewizard_meter" {
						continue
					}

					powerW, batteryPowerW, gridPowerW, energyKwh, soc, err := poller.Poll()
					if err != nil {
						log.Printf("PollerManager: Error polling device %d: %v", id, err)

						pm.cacheMu.Lock()
						pm.deviceCache[id] = DeviceData{
							PowerW:        0,
							BatteryPowerW: 0,
							GridPowerW:    0,
							EnergyKwh:     0,
							Soc:           0,
							Status:        poller.Status(),
							Template:      device.Template,
							HasGridMeter:  device.HasGridMeter,
							HasBattery:    device.HasBattery,
						}
						pm.cacheMu.Unlock()
						continue
					}

					pm.cacheMu.Lock()
					pm.deviceCache[id] = DeviceData{
						PowerW:        powerW,
						BatteryPowerW: batteryPowerW,
						GridPowerW:    gridPowerW,
						EnergyKwh:     energyKwh,
						Soc:           soc,
						Status:        poller.Status(),
						Template:      device.Template,
						HasGridMeter:  device.HasGridMeter,
						HasBattery:    device.HasBattery,
					}
					pm.cacheMu.Unlock()

					// Buffer measurement
					pm.bufferMu.Lock()
					pm.buffer = append(pm.buffer, BufferedMeasurement{
						DeviceID:  id,
						PowerW:    powerW,
						EnergyKwh: energyKwh,
					})
					pm.bufferMu.Unlock()
				}
				pm.mu.Unlock()

				pm.broadcastState()

			case <-pm.stopCh:
				log.Println("PollerManager: Polling stopped")
				return
			}
		}
	}()

	// Start flush loop
	go func() {
		flushTicker := time.NewTicker(1 * time.Minute)
		defer flushTicker.Stop()

		for {
			select {
			case <-flushTicker.C:
				pm.flushBuffer()
			case <-pm.stopCh:
				log.Println("PollerManager: DB Flush stopped")
				return
			}
		}
	}()
}

func (pm *PollerManager) broadcastState() {
	pm.cacheMu.Lock()
	defer pm.cacheMu.Unlock()

	var totalGrid *float64
	var totalSolar *float64
	var totalBattery *float64
	var totalBatterySoc *float64
	var totalEvCharger *float64

	deviceHealth := make(map[int]string)

	for id, data := range pm.deviceCache {
		deviceHealth[id] = data.Status

		switch data.Template {
		case "demo_battery":
			if totalBattery == nil {
				v := 0.0
				totalBattery = &v
			}
			*totalBattery += data.BatteryPowerW

			if totalBatterySoc == nil || *totalBatterySoc == 0 {
				soc := data.Soc
				totalBatterySoc = &soc
			}
		case "huawei_inverter", "solis_inverter", "sma_inverter", "demo_inverter":
			if totalSolar == nil {
				v := 0.0
				totalSolar = &v
			}
			*totalSolar += data.PowerW

			if totalBattery == nil {
				v := 0.0
				totalBattery = &v
			}
			*totalBattery += data.BatteryPowerW

			// For simplicity, take the first battery's SOC or average it if multiple exist
			// Let's just take the most recent / highest non-zero one for now or first found
			if data.HasBattery {
				if totalBatterySoc == nil || *totalBatterySoc == 0 {
					soc := data.Soc
					totalBatterySoc = &soc
				}
			}

			if data.Template == "huawei_inverter" && data.HasGridMeter {
				if totalGrid == nil {
					v := 0.0
					totalGrid = &v
				}
				*totalGrid += data.GridPowerW
			}
		case "homewizard_meter", "demo_dongle":
			if totalGrid == nil {
				v := 0.0
				totalGrid = &v
			}
			*totalGrid += data.GridPowerW
		case "raedian_charger", "alfen_charger", "bender_charger", "phoenix_charger", "easee_charger", "peblar_charger", "demo_charger":
			if totalEvCharger == nil {
				v := 0.0
				totalEvCharger = &v
			}
			*totalEvCharger += data.PowerW
		}
	}

	var totalLoad *float64
	if totalGrid != nil || totalSolar != nil || totalBattery != nil {
		v := 0.0
		if totalGrid != nil { v += *totalGrid }
		if totalSolar != nil { v += *totalSolar }
		if totalBattery != nil { v += *totalBattery }
		totalLoad = &v
	}

	peak := GetProjectedQuarterPeakW()

	state := SiteState{
		GridPowerW:            totalGrid,
		SolarPowerW:           totalSolar,
		BatteryPowerW:         totalBattery,
		BatterySoc:            totalBatterySoc,
		TotalLoadW:            totalLoad,
		EvChargerPowerW:       totalEvCharger,
		ProjectedQuarterPeakW: &peak,
		DeviceHealth:          deviceHealth,
	}

	GlobalStateDispatcher.Broadcast(state)
}

func (pm *PollerManager) flushBuffer() {
	pm.bufferMu.Lock()
	if len(pm.buffer) == 0 {
		pm.bufferMu.Unlock()
		return
	}

	// Make a copy and clear buffer
	currentBuffer := make([]BufferedMeasurement, len(pm.buffer))
	copy(currentBuffer, pm.buffer)
	pm.buffer = make([]BufferedMeasurement, 0)
	pm.bufferMu.Unlock()

	// Aggregate averages by device ID
	type sumCount struct {
		SumPowerW float64
		SumEnergy float64
		Count     int
	}
	agg := make(map[int]*sumCount)

	for _, m := range currentBuffer {
		if _, ok := agg[m.DeviceID]; !ok {
			agg[m.DeviceID] = &sumCount{}
		}
		agg[m.DeviceID].SumPowerW += m.PowerW
		agg[m.DeviceID].SumEnergy += m.EnergyKwh
		agg[m.DeviceID].Count++
	}

	// Transactional insert to limit SD card I/O
	tx, err := db.Begin()
	if err != nil {
		log.Printf("PollerManager DB Flush: Error starting transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO measurements (device_id, power_w, energy_kwh) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		log.Printf("PollerManager DB Flush: Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	for id, data := range agg {
		avgPower := data.SumPowerW / float64(data.Count)
		totalEnergy := data.SumEnergy // energy is incremental, we just sum it for the minute interval

		_, err := stmt.Exec(id, avgPower, totalEnergy)
		if err != nil {
			log.Printf("PollerManager DB Flush: Error executing statement for device %d: %v", id, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("PollerManager DB Flush: Error committing transaction: %v", err)
	}
}

func (pm *PollerManager) Stop() {
	close(pm.stopCh)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, poller := range pm.pollers {
		poller.Close()
	}
	pm.pollers = make(map[int]models.DevicePoller)
}
