package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type GrowattInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "growatt_inverter",
			Name:     "Growatt Inverter",
			Vendor:   "Growatt",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GrowattInverterPoller{Device: device}
		},
	})
}

func (p *GrowattInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] GrowattInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] GrowattInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] GrowattInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *GrowattInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *GrowattInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// PAC (active power) Input Reg 35 or 1000+ depending on model
	// We'll use Input Register 35, 2 words for Growatt MIN/MIC
	powerRegs, err := p.client.ReadRegisters(35, 2, modbus.INPUT_REGISTER)
	if err == nil {
		powerW = float64(uint32(powerRegs[0])<<16|uint32(powerRegs[1])) * 0.1
	}

	// Energy Today (Eac_today) Input Reg 53, 2 words
	energyRegs, err := p.client.ReadRegisters(53, 2, modbus.INPUT_REGISTER)
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.1
	}

	if p.Device.HasBattery {
		// Pbat Input Reg 1009, 2 words
		batRegs, err := p.client.ReadRegisters(1009, 2, modbus.INPUT_REGISTER)
		if err == nil {
			batteryPowerW = float64(int32(uint32(batRegs[0])<<16|uint32(batRegs[1]))) * 0.1
		}
		// SOC Input Reg 1014
		socRegs, err := p.client.ReadRegisters(1014, 1, modbus.INPUT_REGISTER)
		if err == nil {
			soc = float64(socRegs[0])
		}
	}

	if p.Device.HasGridMeter {
		// Pgrid Input Reg 1012, 2 words
		gridRegs, err := p.client.ReadRegisters(1012, 2, modbus.INPUT_REGISTER)
		if err == nil {
			gridPowerW = float64(int32(uint32(gridRegs[0])<<16|uint32(gridRegs[1]))) * 0.1
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *GrowattInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *GrowattInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
