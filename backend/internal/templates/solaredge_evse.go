package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SolarEdgeEvsePoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "solaredge_evse",
			Name:     "SolarEdge EV Charger",
			Vendor:   "SolarEdge",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SolarEdgeEvsePoller{Device: device}
		},
	})
}

func (p *SolarEdgeEvsePoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SolarEdgeEvsePoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SolarEdgeEvsePoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SolarEdgeEvsePoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SolarEdgeEvsePoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SolarEdgeEvsePoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	// SolarEdge EV charger registers (SunSpec style EVSE)
	// Usually around 40000+, we will use standard EVSE modbus mapping based on SolarEdge Modbus specs

	powerRegs, err := p.client.ReadRegisters(40083, 1, modbus.HOLDING_REGISTER) // Total active power
	powerW := 0.0
	if err == nil {
		scaleRegs, err := p.client.ReadRegisters(40084, 1, modbus.HOLDING_REGISTER)
		if err == nil && len(scaleRegs) > 0 {
			scale := float64(int16(scaleRegs[0]))
			powerW = float64(int16(powerRegs[0]))
			for i := 0; i < int(scale); i++ {
				powerW *= 10
			}
			for i := 0; i > int(scale); i-- {
				powerW /= 10
			}
		}
	}

	energyRegs, err := p.client.ReadRegisters(40093, 2, modbus.HOLDING_REGISTER) // Total Energy
	energyKwh := 0.0
	if err == nil {
		scaleRegs, err := p.client.ReadRegisters(40095, 1, modbus.HOLDING_REGISTER)
		if err == nil && len(scaleRegs) > 0 {
			scale := float64(int16(scaleRegs[0]))
			energyWh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))

			for i := 0; i < int(scale); i++ {
				energyWh *= 10
			}
			for i := 0; i > int(scale); i-- {
				energyWh /= 10
			}
			energyKwh = energyWh / 1000.0
		}
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *SolarEdgeEvsePoller) GetDevice() models.Device {
	return p.Device
}

func (p *SolarEdgeEvsePoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	return p.client.WriteRegister(40000, uint16(amps)) // Generic EVSE mapping write register
}

func (p *SolarEdgeEvsePoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
