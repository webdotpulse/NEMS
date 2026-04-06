package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SolarEdgeInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "solaredge_inverter",
			Name:     "SolarEdge Inverter",
			Vendor:   "SolarEdge",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SolarEdgeInverterPoller{Device: device}
		},
	})
}

func (p *SolarEdgeInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SolarEdgeInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SolarEdgeInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SolarEdgeInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SolarEdgeInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SolarEdgeInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// Active Power (40083) and Scale (40084)
	powerRegs, err := p.client.ReadRegisters(40083, 2, modbus.HOLDING_REGISTER)
	if err == nil {
		scale := float64(int16(powerRegs[1]))
		powerW = float64(int16(powerRegs[0]))
		for i := 0; i < int(scale); i++ {
			powerW *= 10
		}
		for i := 0; i > int(scale); i-- {
			powerW /= 10
		}
	}

	// Energy (40093) and Scale (40095)
	energyRegs, err := p.client.ReadRegisters(40093, 3, modbus.HOLDING_REGISTER)
	if err == nil {
		scale := float64(int16(energyRegs[2]))
		energyWh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))
		for i := 0; i < int(scale); i++ {
			energyWh *= 10
		}
		for i := 0; i > int(scale); i-- {
			energyWh /= 10
		}
		energyKwh = energyWh / 1000.0
	}

	if p.Device.HasBattery {
		// Battery Power (57716, float32)
		batRegs, err := p.client.ReadRegisters(57716, 2, modbus.HOLDING_REGISTER)
		if err == nil {
			// EVCC uses IEEE754 float for SolarEdge battery power
			batteryPowerW = float64(decodeFloat32(batRegs))
		}
		// SOC (57710, float32)
		socRegs, err := p.client.ReadRegisters(57710, 2, modbus.HOLDING_REGISTER)
		if err == nil {
			soc = float64(decodeFloat32(socRegs))
		}
	}

	if p.Device.HasGridMeter {
		// Meter Power (40206) and Scale (40210)
		gridRegs, err := p.client.ReadRegisters(40206, 5, modbus.HOLDING_REGISTER)
		if err == nil {
			scale := float64(int16(gridRegs[4]))
			gridPowerW = float64(int16(gridRegs[0]))
			for i := 0; i < int(scale); i++ {
				gridPowerW *= 10
			}
			for i := 0; i > int(scale); i-- {
				gridPowerW /= 10
			}
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *SolarEdgeInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SolarEdgeInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
