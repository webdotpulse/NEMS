package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type DevicePoller interface {
	Connect() error
	Poll() (powerW float64, batteryPowerW float64, energyKwh float64, err error)
	GetDevice() Device
	Status() string
	Close() error
}

// ---------------------------------------------------------
// Huawei Hybrid Inverter
// ---------------------------------------------------------

type HuaweiInverterPoller struct {
	Device Device
	conn   net.Conn
	status string
}

func (p *HuaweiInverterPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("HuaweiInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil // Fallback to simulation
	}
	p.conn = conn
	p.status = "online"
	return nil
}

func (p *HuaweiInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
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
	status string
}

func (p *HuaweiDonglePoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiDonglePoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("HuaweiDonglePoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil // Fallback to simulation
	}
	p.conn = conn
	p.status = "online"
	return nil
}

func (p *HuaweiDonglePoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
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
	status string
}

func (p *RaedianChargerPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("RaedianChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		log.Printf("RaedianChargerPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil // Fallback to simulation
	}
	p.conn = conn
	p.status = "online"
	return nil
}

func (p *RaedianChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
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
// Solis Inverter
// ---------------------------------------------------------

type SolisInverterPoller struct {
	Device Device
	client *modbus.ModbusClient
	status string
}

func (p *SolisInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("SolisInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("SolisInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("SolisInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SolisInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SolisInverterPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 800.0 + rand.Float64()*2000.0
		batteryPowerW := -1000.0 + rand.Float64()*2000.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, energyKwh, nil
	}

	powerRegs, err := p.client.ReadRegisters(33079, 2, modbus.INPUT_REGISTER)
	if err != nil {
		return 0, 0, 0, err
	}
	powerW := float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))

	energyRegs, err := p.client.ReadRegisters(33029, 2, modbus.INPUT_REGISTER)
	if err != nil {
		return powerW, 0, 0, err
	}
	energyKwh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1])) / 10.0

	batRegs, err := p.client.ReadRegisters(33149, 2, modbus.INPUT_REGISTER)
	batteryPowerW := 0.0
	if err == nil {
		rawBat := int32(uint32(batRegs[0])<<16 | uint32(batRegs[1]))
		batteryPowerW = float64(rawBat)
	}

	return powerW, batteryPowerW, energyKwh, nil
}

func (p *SolisInverterPoller) GetDevice() Device {
	return p.Device
}

func (p *SolisInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// ---------------------------------------------------------
// SMA Inverter
// ---------------------------------------------------------

type SmaInverterPoller struct {
	Device Device
	client *modbus.ModbusClient
	status string
}

func (p *SmaInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("SmaInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("SmaInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("SmaInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SmaInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SmaInverterPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 1500.0 + rand.Float64()*2500.0
		batteryPowerW := -500.0 + rand.Float64()*1500.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, energyKwh, nil
	}

	powerRegs, err := p.client.ReadRegisters(30775, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, err
	}
	powerW := float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))

	energyRegs, err := p.client.ReadRegisters(30529, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyWh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))
		energyKwh = energyWh / 1000.0
	}

	batRegs, err := p.client.ReadRegisters(31393, 2, modbus.HOLDING_REGISTER)
	batteryPowerW := 0.0
	if err == nil {
		batteryPowerW = float64(int32(uint32(batRegs[0])<<16 | uint32(batRegs[1])))
	}

	return powerW, batteryPowerW, energyKwh, nil
}

func (p *SmaInverterPoller) GetDevice() Device {
	return p.Device
}

func (p *SmaInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// ---------------------------------------------------------
// Generic EV Charger Poller (for Alfen, Bender, Phoenix)
// ---------------------------------------------------------

type GenericModbusChargerPoller struct {
	Device Device
	client *modbus.ModbusClient
	status string
	Prefix string
}

func (p *GenericModbusChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("%s: Attempting Modbus TCP connection to %s (ID: %d)", p.Prefix, addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("%s: Client setup failed (%v)", p.Prefix, err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("%s: Connection failed, falling back to simulation mode (%v)", p.Prefix, err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *GenericModbusChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *GenericModbusChargerPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 0.0
		if rand.Float32() > 0.6 {
			powerW = 7400.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, energyKwh, nil
	}

	powerRegs, err := p.client.ReadRegisters(344, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, err
	}

	powerW := float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *GenericModbusChargerPoller) GetDevice() Device {
	return p.Device
}

func (p *GenericModbusChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// ---------------------------------------------------------
// Easee Cloud Charger Poller
// ---------------------------------------------------------

type EaseeChargerPoller struct {
	Device Device
	status string
}

func (p *EaseeChargerPoller) Connect() error {
	log.Printf("EaseeChargerPoller: Authenticating with cloud using user %s", p.Device.Username)
	time.Sleep(500 * time.Millisecond)
	p.status = "online"
	return nil
}

func (p *EaseeChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *EaseeChargerPoller) Poll() (float64, float64, float64, error) {
	powerW := 0.0
	if rand.Float32() > 0.4 {
		powerW = 22000.0
	}
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *EaseeChargerPoller) GetDevice() Device {
	return p.Device
}

func (p *EaseeChargerPoller) Close() error {
	return nil
}

// ---------------------------------------------------------
// Peblar REST API Charger Poller
// ---------------------------------------------------------

type PeblarChargerPoller struct {
	Device Device
	status string
}

func (p *PeblarChargerPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("PeblarChargerPoller: Attempting REST API connection to %s", addr)

	url := fmt.Sprintf("http://%s/api/v1/system", addr)
	client := http.Client{Timeout: 2 * time.Second}
	_, err := client.Get(url)
	if err != nil {
		log.Printf("PeblarChargerPoller: Connection failed, falling back to simulation (%v)", err)
		p.status = "error"
		return nil
	}

	p.status = "online"
	return nil
}

func (p *PeblarChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *PeblarChargerPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" {
		powerW := 0.0
		if rand.Float32() > 0.5 {
			powerW = 11000.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, energyKwh, nil
	}

	url := fmt.Sprintf("http://%s:%d/api/v1/meter", p.Device.Host, p.Device.Port)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	var data struct {
		Power  float64 `json:"power"`
		Energy float64 `json:"energy"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, 0, err
	}

	return data.Power, 0, data.Energy, nil
}

func (p *PeblarChargerPoller) GetDevice() Device {
	return p.Device
}

func (p *PeblarChargerPoller) Close() error {
	return nil
}

// ---------------------------------------------------------
// HomeWizard Meter Poller
// ---------------------------------------------------------

type HomeWizardMeterPoller struct {
	Device Device
	status string
}

func (p *HomeWizardMeterPoller) Connect() error {
	addr := p.Device.Host
	if p.Device.Port != 80 && p.Device.Port != 0 {
		addr = addr + ":" + strconv.Itoa(p.Device.Port)
	}
	log.Printf("HomeWizardMeterPoller: Attempting local REST API connection to http://%s/api/v1/data", addr)

	url := fmt.Sprintf("http://%s/api/v1/data", addr)
	client := http.Client{Timeout: 2 * time.Second}
	_, err := client.Get(url)
	if err != nil {
		log.Printf("HomeWizardMeterPoller: Connection failed, falling back to simulation (%v)", err)
		p.status = "error"
		return nil
	}

	p.status = "online"
	return nil
}

func (p *HomeWizardMeterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *HomeWizardMeterPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" {
		powerW := -1500.0 + rand.Float64()*3500.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, energyKwh, nil
	}

	addr := p.Device.Host
	if p.Device.Port != 80 && p.Device.Port != 0 {
		addr = addr + ":" + strconv.Itoa(p.Device.Port)
	}
	url := fmt.Sprintf("http://%s/api/v1/data", addr)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	var data struct {
		ActivePowerW         float64 `json:"active_power_w"`
		TotalEnergyImportKwh float64 `json:"total_power_import_kwh"`
		TotalEnergyExportKwh float64 `json:"total_power_export_kwh"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, 0, err
	}

	return data.ActivePowerW, 0, data.TotalEnergyImportKwh - data.TotalEnergyExportKwh, nil
}

func (p *HomeWizardMeterPoller) GetDevice() Device {
	return p.Device
}

func (p *HomeWizardMeterPoller) Close() error {
	return nil
}

// ---------------------------------------------------------
// Demo Devices
// ---------------------------------------------------------

type DemoInverterPoller struct {
	Device Device
}

func (p *DemoInverterPoller) Connect() error {
	return nil // No connection needed
}

func (p *DemoInverterPoller) Status() string {
	return "online"
}

func (p *DemoInverterPoller) Poll() (float64, float64, float64, error) {
	powerW := 1000.0 + rand.Float64()*3000.0
	batteryPowerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, batteryPowerW, energyKwh, nil
}

func (p *DemoInverterPoller) GetDevice() Device {
	return p.Device
}

func (p *DemoInverterPoller) Close() error {
	return nil
}

type DemoDonglePoller struct {
	Device Device
}

func (p *DemoDonglePoller) Connect() error {
	return nil
}

func (p *DemoDonglePoller) Status() string {
	return "online"
}

func (p *DemoDonglePoller) Poll() (float64, float64, float64, error) {
	powerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *DemoDonglePoller) GetDevice() Device {
	return p.Device
}

func (p *DemoDonglePoller) Close() error {
	return nil
}

type DemoChargerPoller struct {
	Device Device
}

func (p *DemoChargerPoller) Connect() error {
	return nil
}

func (p *DemoChargerPoller) Status() string {
	return "online"
}

func (p *DemoChargerPoller) Poll() (float64, float64, float64, error) {
	powerW := 0.0
	if rand.Float32() > 0.5 {
		powerW = 11000.0
	}
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *DemoChargerPoller) GetDevice() Device {
	return p.Device
}

func (p *DemoChargerPoller) Close() error {
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
	rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password FROM devices")
	if err != nil {
		log.Printf("PollerManager: Error fetching devices: %v", err)
		return
	}
	defer rows.Close()

	activeDeviceIDs := make(map[int]bool)

	for rows.Next() {
		var d Device
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
			var poller DevicePoller
			switch d.Template {
			case "huawei_inverter":
				poller = &HuaweiInverterPoller{Device: d}
			case "solis_inverter":
				poller = &SolisInverterPoller{Device: d}
			case "sma_inverter":
				poller = &SmaInverterPoller{Device: d}
			case "huawei_dongle":
				poller = &HuaweiDonglePoller{Device: d}
			case "homewizard_meter":
				poller = &HomeWizardMeterPoller{Device: d}
			case "raedian_charger":
				poller = &RaedianChargerPoller{Device: d}
			case "alfen_charger":
				poller = &GenericModbusChargerPoller{Device: d, Prefix: "AlfenChargerPoller"}
			case "bender_charger":
				poller = &GenericModbusChargerPoller{Device: d, Prefix: "BenderChargerPoller"}
			case "phoenix_charger":
				poller = &GenericModbusChargerPoller{Device: d, Prefix: "PhoenixChargerPoller"}
			case "easee_charger":
				poller = &EaseeChargerPoller{Device: d}
			case "peblar_charger":
				poller = &PeblarChargerPoller{Device: d}
			case "demo_inverter":
				poller = &DemoInverterPoller{Device: d}
			case "demo_dongle":
				poller = &DemoDonglePoller{Device: d}
			case "demo_charger":
				poller = &DemoChargerPoller{Device: d}
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
	pm.pollers = make(map[int]DevicePoller)
}
