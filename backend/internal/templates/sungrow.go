package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SungrowInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "sungrow_inverter",
			Name:     "Sungrow Inverter",
			Vendor:   "Sungrow",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SungrowInverterPoller{Device: device}
		},
	})
}

func (p *SungrowInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SungrowInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SungrowInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SungrowInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SungrowInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SungrowInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// Active Power (13007)
	powerRegs, err := p.client.ReadRegisters(13007, 2, modbus.INPUT_REGISTER)
	if err == nil {
		powerW = float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))
	}

	// Energy Yield Today (13001)
	energyRegs, err := p.client.ReadRegisters(13001, 2, modbus.INPUT_REGISTER)
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.1
	}

	if p.Device.HasBattery {
		// Battery Power (13021)
		batRegs, err := p.client.ReadRegisters(13021, 1, modbus.INPUT_REGISTER)
		if err == nil {
			batteryPowerW = float64(int16(batRegs[0])) * 10.0 // Sungrow gives tens of watts or sometimes W based on model, we use W
		}
		// SOC (13022)
		socRegs, err := p.client.ReadRegisters(13022, 1, modbus.INPUT_REGISTER)
		if err == nil {
			soc = float64(socRegs[0]) * 0.1
		}
	}

	if p.Device.HasGridMeter {
		// Grid Power (13009)
		gridRegs, err := p.client.ReadRegisters(13009, 2, modbus.INPUT_REGISTER)
		if err == nil {
			gridPowerW = float64(int32(uint32(gridRegs[0])<<16 | uint32(gridRegs[1])))
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *SungrowInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SungrowInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
