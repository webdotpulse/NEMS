package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SofarSolarInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "sofarsolar_inverter",
			Name:     "SofarSolar Inverter",
			Vendor:   "SofarSolar",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SofarSolarInverterPoller{Device: device}
		},
	})
}

func (p *SofarSolarInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SofarSolarInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SofarSolarInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SofarSolarInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SofarSolarInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SofarSolarInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// Active Power (0x000C)
	powerRegs, err := p.client.ReadRegisters(0x000C, 1, modbus.HOLDING_REGISTER)
	if err == nil {
		powerW = float64(powerRegs[0]) * 10.0
	}

	// Energy (0x0015, 2)
	energyRegs, err := p.client.ReadRegisters(0x0015, 2, modbus.HOLDING_REGISTER)
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))
	}

	if p.Device.HasBattery {
		// Battery Power (0x020D)
		batRegs, err := p.client.ReadRegisters(0x020D, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			batteryPowerW = float64(int16(batRegs[0])) * 10.0
		}
		// SOC (0x0210)
		socRegs, err := p.client.ReadRegisters(0x0210, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			soc = float64(socRegs[0])
		}
	}

	if p.Device.HasGridMeter {
		// Grid Power (0x0212)
		gridRegs, err := p.client.ReadRegisters(0x0212, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			gridPowerW = float64(int16(gridRegs[0])) * 10.0
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *SofarSolarInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SofarSolarInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
