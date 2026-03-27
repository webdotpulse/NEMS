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
}

var PollerMgr *PollerManager

func InitPollerManager() {
	PollerMgr = &PollerManager{
		pollers: make(map[int]models.DevicePoller),
		stopCh:  make(chan struct{}),
		buffer:  make([]BufferedMeasurement, 0),
	}
}

func (pm *PollerManager) SyncDevices() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	log.Println("PollerManager: Syncing devices...")

	// Fetch current devices from DB
	rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password FROM devices")
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
		if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password); err != nil {
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
			log.Printf("PollerManager: Removed poller for device %d", id)
		}
	}
}

func (pm *PollerManager) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pm.mu.Lock()

				var totalGrid *float64
				var totalSolar *float64
				var totalBattery *float64
				var totalEvCharger *float64

				deviceHealth := make(map[int]string)

				for id, poller := range pm.pollers {
					deviceHealth[id] = poller.Status()

					powerW, batteryPowerW, energyKwh, err := poller.Poll()
					if err != nil {
						log.Printf("PollerManager: Error polling device %d: %v", id, err)
						continue
					}

					device := poller.GetDevice()
					switch device.Template {
					case "huawei_inverter", "solis_inverter", "sma_inverter", "demo_inverter":
						if totalSolar == nil {
							v := 0.0
							totalSolar = &v
						}
						*totalSolar += powerW

						if totalBattery == nil {
							v := 0.0
							totalBattery = &v
						}
						*totalBattery += batteryPowerW
					case "huawei_dongle", "homewizard_meter", "demo_dongle":
						if totalGrid == nil {
							v := 0.0
							totalGrid = &v
						}
						*totalGrid += powerW
					case "raedian_charger", "alfen_charger", "bender_charger", "phoenix_charger", "easee_charger", "peblar_charger", "demo_charger":
						if totalEvCharger == nil {
							v := 0.0
							totalEvCharger = &v
						}
						*totalEvCharger += powerW
					}

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

				// Total Load = Grid Power (imported positive) + Solar + Battery (discharging positive)
				// Dongle Import is positive.
				// Inverter Solar is positive.
				// Inverter Battery Discharging is positive.
				// This assumes powerW from Dongle is positive for IMPORT and negative for EXPORT.
				// Wait, the Dongle simulation says "positive = import, negative = export".

				var totalLoad *float64

				// Calculate total load only if we have at least one valid measurement
				if totalGrid != nil || totalSolar != nil || totalBattery != nil {
					v := 0.0
					if totalGrid != nil { v += *totalGrid }
					if totalSolar != nil { v += *totalSolar }
					if totalBattery != nil { v += *totalBattery }
					totalLoad = &v
				}

				state := SiteState{
					GridPowerW:      totalGrid,
					SolarPowerW:     totalSolar,
					BatteryPowerW:   totalBattery,
					TotalLoadW:      totalLoad,
					EvChargerPowerW: totalEvCharger,
					DeviceHealth:    deviceHealth,
				}

				GlobalStateDispatcher.Broadcast(state)

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
