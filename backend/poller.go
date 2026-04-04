package main

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"sync/atomic"
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
	Category      string
	HasGridMeter  bool
	HasBattery    bool
}

type BufferedMeasurement struct {
	DeviceID      int
	PowerW        float64
	BatteryPowerW float64
	GridPowerW    float64
	EnergyKwh     float64
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

// InitPollerManager initializes the global PollerMgr instance.
func InitPollerManager() {
	PollerMgr = &PollerManager{
		pollers: make(map[int]models.DevicePoller),
		stopCh:  make(chan struct{}),
		buffer:  make([]BufferedMeasurement, 0),
		deviceCache: make(map[int]DeviceData),
	}
}

// SyncDevices fetches all active devices from the database and ensures
// the correct DevicePoller templates are running. It creates new pollers
// and removes outdated ones.
func (pm *PollerManager) SyncDevices() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	log.Println("[INFO] PollerManager: Syncing devices...")

	// Fetch current devices from DB
	rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, ocpp_proxy_url FROM devices")
	if err != nil {
		log.Printf("[ERROR] PollerManager: Error fetching devices: %v", err)
		return
	}
	defer rows.Close()

	activeDeviceIDs := make(map[int]bool)

	for rows.Next() {
		var d models.Device
		var username sql.NullString
		var password sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password, &d.HasGridMeter, &d.HasBattery, &d.BatteryCapacity); err != nil {
			log.Printf("[ERROR] PollerManager: Error scanning device: %v", err)
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
				log.Printf("[INFO] PollerManager: Unknown template %s for device %d", d.Template, d.ID)
				continue
			}

			err := poller.Connect()
			if err != nil {
				log.Printf("[ERROR] PollerManager: Failed to connect device %d: %v", d.ID, err)
			}
			pm.pollers[d.ID] = poller
			log.Printf("[INFO] PollerManager: Added poller for device %d (%s)", d.ID, d.Name)
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
			log.Printf("[INFO] PollerManager: Removed poller for device %d", id)
		}
	}
}

// Start begins the asynchronous polling loops.
// It uses a fast 1-second ticker for smart meters (e.g. HomeWizard) and
// a 5-second ticker for standard devices. It also starts a background
// goroutine to flush buffered data to SQLite every minute.
func (pm *PollerManager) Start() {
	go func() {
		fastTicker := time.NewTicker(1 * time.Second)
		defer fastTicker.Stop()

		lastPolled := make(map[int]time.Time)

		for {
			select {
			case <-fastTicker.C:
				now := time.Now()
				pollersCopy := pm.GetPollers()
				var polledAny atomic.Bool
				var wg sync.WaitGroup

				for id, poller := range pollersCopy {
					device := poller.GetDevice()
					category := templates.GetCategory(device.Template)

					interval := device.PollInterval
					if interval <= 0 {
						if category == "meter" {
							interval = 1
						} else {
							interval = 5
						}
					}

					if last, ok := lastPolled[id]; ok {
						if now.Sub(last) < time.Duration(interval)*time.Second {
							continue
						}
					}

					lastPolled[id] = now
					polledAny.Store(true)

					wg.Add(1)
					go func(id int, poller models.DevicePoller, device models.Device, category string) {
						defer wg.Done()
						powerW, batteryPowerW, gridPowerW, energyKwh, soc, err := poller.Poll()
						if err != nil {
							log.Printf("[ERROR] PollerManager: Error polling device %d: %v", id, err)

							errStr := err.Error()
							if strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "EOF") || strings.Contains(errStr, "connection reset") {
								log.Printf("[INFO] PollerManager: Connection drop detected for device %d, attempting to reconnect...", id)
								poller.Close()
								if connErr := poller.Connect(); connErr != nil {
									log.Printf("[ERROR] PollerManager: Reconnect failed for device %d: %v", id, connErr)
								}
							}

							pm.cacheMu.Lock()
							oldData, exists := pm.deviceCache[id]
							if exists {
								oldData.Status = poller.Status()
								pm.deviceCache[id] = oldData
							} else {
								pm.deviceCache[id] = DeviceData{
									PowerW:        0,
									BatteryPowerW: 0,
									GridPowerW:    0,
									EnergyKwh:     0,
									Soc:           0,
									Status:        poller.Status(),
									Template:      device.Template,
									Category:      category,
									HasGridMeter:  device.HasGridMeter,
									HasBattery:    device.HasBattery,
								}
							}
							pm.cacheMu.Unlock()
							return
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
							Category:      category,
							HasGridMeter:  device.HasGridMeter,
							HasBattery:    device.HasBattery,
						}
						pm.cacheMu.Unlock()

						// Buffer measurement
						pm.bufferMu.Lock()
						pm.buffer = append(pm.buffer, BufferedMeasurement{
							DeviceID:      id,
							PowerW:        powerW,
							BatteryPowerW: batteryPowerW,
							GridPowerW:    gridPowerW,
							EnergyKwh:     energyKwh,
						})
						pm.bufferMu.Unlock()
					}(id, poller, device, category)
				}
				wg.Wait()

				if polledAny.Load() {
					pm.broadcastState()
				}

			case <-pm.stopCh:
				log.Println("[INFO] PollerManager: Polling stopped")
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
				log.Println("[INFO] PollerManager: DB Flush stopped")
				return
			}
		}
	}()
}

// broadcastState aggregates all currently cached device data into a single
// SiteState struct and broadcasts it to any connected SSE clients via GlobalStateDispatcher.
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

		switch data.Category {
		case "battery":
			if totalBattery == nil {
				v := 0.0
				totalBattery = &v
			}
			*totalBattery += data.BatteryPowerW

			if totalBatterySoc == nil || *totalBatterySoc == 0 {
				soc := data.Soc
				totalBatterySoc = &soc
			}
		case "inverter":
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

			if data.HasGridMeter {
				if totalGrid == nil {
					v := 0.0
					totalGrid = &v
				}
				*totalGrid += data.GridPowerW
			}
		case "meter":
			if totalGrid == nil {
				v := 0.0
				totalGrid = &v
			}
			*totalGrid += data.GridPowerW
		case "charger":
			if totalEvCharger == nil {
				v := 0.0
				totalEvCharger = &v
			}
			*totalEvCharger += data.PowerW
		}
	}

	var totalLoad *float64
	if totalGrid != nil {
		v := *totalGrid
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

// flushBuffer aggregates the 1-second/5-second in-memory power measurements
// into a 1-minute average and flushes them to the SQLite `measurements` table
// using a single database transaction to limit SD card wear.
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
		SumPowerW        float64
		SumBatteryPowerW float64
		SumGridPowerW    float64
		LastEnergy       float64
		Count            int
	}
	agg := make(map[int]*sumCount)

	for _, m := range currentBuffer {
		if _, ok := agg[m.DeviceID]; !ok {
			agg[m.DeviceID] = &sumCount{}
		}
		agg[m.DeviceID].SumPowerW += m.PowerW
		agg[m.DeviceID].SumBatteryPowerW += m.BatteryPowerW
		agg[m.DeviceID].SumGridPowerW += m.GridPowerW
		agg[m.DeviceID].LastEnergy = m.EnergyKwh
		agg[m.DeviceID].Count++
	}

	// Transactional insert to limit SD card I/O
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] PollerManager DB Flush: Error starting transaction: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO measurements (device_id, power_w, battery_power_w, grid_power_w, energy_kwh) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] PollerManager DB Flush: Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	for id, data := range agg {
		avgPower := data.SumPowerW / float64(data.Count)
		avgBatteryPower := data.SumBatteryPowerW / float64(data.Count)
		avgGridPower := data.SumGridPowerW / float64(data.Count)
		totalEnergy := data.LastEnergy // store the latest cumulative energy read in this minute interval

		_, err := stmt.Exec(id, avgPower, avgBatteryPower, avgGridPower, totalEnergy)
		if err != nil {
			log.Printf("[ERROR] PollerManager DB Flush: Error executing statement for device %d: %v", id, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] PollerManager DB Flush: Error committing transaction: %v", err)
	}
}

// Stop gracefully shuts down all active pollers and stops the background sync tasks.
func (pm *PollerManager) Stop() {
	close(pm.stopCh)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, poller := range pm.pollers {
		poller.Close()
	}
	pm.pollers = make(map[int]models.DevicePoller)
}
