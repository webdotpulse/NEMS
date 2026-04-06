package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SolaxInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "solax_inverter",
			Name:     "SolaX Inverter",
			Vendor:   "SolaX",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SolaxInverterPoller{Device: device}
		},
	})
}

func (p *SolaxInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SolaxInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SolaxInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SolaxInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SolaxInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SolaxInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// Inverter Active Power (11)
	powerRegs, err := p.client.ReadRegisters(11, 2, modbus.INPUT_REGISTER)
	if err == nil {
		powerW = float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))
	}

	// Energy (80)
	energyRegs, err := p.client.ReadRegisters(80, 2, modbus.INPUT_REGISTER)
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.1
	}

	if p.Device.HasBattery {
		// Battery Power (22)
		batRegs, err := p.client.ReadRegisters(22, 1, modbus.INPUT_REGISTER)
		if err == nil {
			batteryPowerW = float64(int16(batRegs[0]))
		}
		// SOC (28)
		socRegs, err := p.client.ReadRegisters(28, 1, modbus.INPUT_REGISTER)
		if err == nil {
			soc = float64(socRegs[0])
		}
	}

	if p.Device.HasGridMeter {
		// Grid Power (70)
		gridRegs, err := p.client.ReadRegisters(70, 2, modbus.INPUT_REGISTER)
		if err == nil {
			gridPowerW = float64(int32(uint32(gridRegs[0])<<16 | uint32(gridRegs[1])))
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *SolaxInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SolaxInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
