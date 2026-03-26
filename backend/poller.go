package main

import (
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type DevicePoller interface {
	Connect() error
	Poll() (powerW float64, batteryPowerW float64, energyKwh float64, err error)
	GetDevice() Device
	Close() error
}

// ---------------------------------------------------------
// Huawei Hybrid Inverter
// ---------------------------------------------------------

type HuaweiInverterPoller struct {
	Device Device
	conn   net.Conn
}

func (p *HuaweiInverterPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("HuaweiInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		return nil // Fallback to simulation
	}
	p.conn = conn
	return nil
}

func (p *HuaweiInverterPoller) Poll() (float64, float64, float64, error) {
	// Simulate typical inverter output
	powerW := 1000.0 + rand.Float64()*3000.0
	// Simulate battery charge/discharge (negative = charging, positive = discharging)
	batteryPowerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0 // rough incremental energy for 5s
	return powerW, batteryPowerW, energyKwh, nil
}

func (p *HuaweiInverterPoller) GetDevice() Device {
	return p.Device
}

func (p *HuaweiInverterPoller) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// ---------------------------------------------------------
// Huawei Dongle Power Sensor
// ---------------------------------------------------------

type HuaweiDonglePoller struct {
	Device Device
	conn   net.Conn
}

func (p *HuaweiDonglePoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiDonglePoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("HuaweiDonglePoller: Connection failed, falling back to simulation mode (%v)", err)
		return nil // Fallback to simulation
	}
	p.conn = conn
	return nil
}

func (p *HuaweiDonglePoller) Poll() (float64, float64, float64, error) {
	// Simulate grid meter reading (positive = import, negative = export)
	powerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *HuaweiDonglePoller) GetDevice() Device {
	return p.Device
}

func (p *HuaweiDonglePoller) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// ---------------------------------------------------------
// Raedian EV Charger
// ---------------------------------------------------------

type RaedianChargerPoller struct {
	Device Device
	conn   net.Conn
}

func (p *RaedianChargerPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("RaedianChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("RaedianChargerPoller: Connection failed, falling back to simulation mode (%v)", err)
		return nil // Fallback to simulation
	}
	p.conn = conn
	return nil
}

func (p *RaedianChargerPoller) Poll() (float64, float64, float64, error) {
	// Simulate charging
	powerW := 0.0
	if rand.Float32() > 0.5 {
		powerW = 11000.0 // 11kW charging
	}
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *RaedianChargerPoller) GetDevice() Device {
	return p.Device
}

func (p *RaedianChargerPoller) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// ---------------------------------------------------------
// Poller Manager
// ---------------------------------------------------------

type BufferedMeasurement struct {
	DeviceID  int
	PowerW    float64
	EnergyKwh float64
}

type PollerManager struct {
	pollers map[int]DevicePoller
	mu      sync.Mutex
	stopCh  chan struct{}

	bufferMu sync.Mutex
	buffer   []BufferedMeasurement
}

var PollerMgr *PollerManager

func InitPollerManager() {
	PollerMgr = &PollerManager{
		pollers: make(map[int]DevicePoller),
		stopCh:  make(chan struct{}),
		buffer:  make([]BufferedMeasurement, 0),
	}
}

func (pm *PollerManager) SyncDevices() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	log.Println("PollerManager: Syncing devices...")

	// Fetch current devices from DB
	rows, err := db.Query("SELECT id, name, template, host, port, modbus_id FROM devices")
	if err != nil {
		log.Printf("PollerManager: Error fetching devices: %v", err)
		return
	}
	defer rows.Close()

	activeDeviceIDs := make(map[int]bool)

	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID); err != nil {
			log.Printf("PollerManager: Error scanning device: %v", err)
			continue
		}
		activeDeviceIDs[d.ID] = true

		// If poller doesn't exist, create it
		if _, exists := pm.pollers[d.ID]; !exists {
			var poller DevicePoller
			switch d.Template {
			case "huawei_inverter":
				poller = &HuaweiInverterPoller{Device: d}
			case "huawei_dongle":
				poller = &HuaweiDonglePoller{Device: d}
			case "raedian_charger":
				poller = &RaedianChargerPoller{Device: d}
			default:
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

				var totalGrid float64
				var totalSolar float64
				var totalBattery float64

				for id, poller := range pm.pollers {
					powerW, batteryPowerW, energyKwh, err := poller.Poll()
					if err != nil {
						log.Printf("PollerManager: Error polling device %d: %v", id, err)
						continue
					}

					device := poller.GetDevice()
					switch device.Template {
					case "huawei_inverter":
						totalSolar += powerW
						totalBattery += batteryPowerW
					case "huawei_dongle":
						totalGrid += powerW
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

				totalLoad := totalGrid + totalSolar + totalBattery

				state := SiteState{
					GridPowerW:    totalGrid,
					SolarPowerW:   totalSolar,
					BatteryPowerW: totalBattery,
					TotalLoadW:    totalLoad,
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
	} else {
		log.Printf("PollerManager DB Flush: Flushed %d aggregated measurements to DB", len(agg))
	}
}

func (pm *PollerManager) Stop() {
	close(pm.stopCh)
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, poller := range pm.pollers {
		poller.Close()
	}
	pm.pollers = make(map[int]DevicePoller)
}
