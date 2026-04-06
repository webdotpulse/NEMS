package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type GoodWeInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "goodwe_inverter",
			Name:     "GoodWe Inverter",
			Vendor:   "GoodWe",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GoodWeInverterPoller{Device: device}
		},
	})
}

func (p *GoodWeInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] GoodWeInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] GoodWeInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] GoodWeInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *GoodWeInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *GoodWeInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// PV Power (35105, 2)
	powerRegs, err := p.client.ReadRegisters(35105, 2, modbus.HOLDING_REGISTER)
	if err == nil {
		powerW = float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))
	}

	// Total Energy (35191, 2)
	energyRegs, err := p.client.ReadRegisters(35191, 2, modbus.HOLDING_REGISTER)
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.1
	}

	if p.Device.HasBattery {
		// Battery Power (35183)
		batRegs, err := p.client.ReadRegisters(35183, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			batteryPowerW = float64(int16(batRegs[0]))
		}
		// SOC (37007)
		socRegs, err := p.client.ReadRegisters(37007, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			soc = float64(socRegs[0])
		}
	}

	if p.Device.HasGridMeter {
		// Grid Power (35140, 2)
		gridRegs, err := p.client.ReadRegisters(35140, 2, modbus.HOLDING_REGISTER)
		if err == nil {
			gridPowerW = float64(int32(uint32(gridRegs[0])<<16 | uint32(gridRegs[1])))
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *GoodWeInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *GoodWeInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
