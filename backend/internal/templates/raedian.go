package templates

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type RaedianChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "raedian_charger",
			Name: "Raedian EV Charger",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &RaedianChargerPoller{Device: device}
		},
	})
}

func (p *RaedianChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("RaedianChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("RaedianChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("RaedianChargerPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *RaedianChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *RaedianChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 0.0
		if rand.Float32() > 0.5 {
			powerW = 11000.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, 0, energyKwh, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(32796, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	powerW := float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))

	energyRegs, err := p.client.ReadRegisters(32798, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyWh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))
		energyKwh = energyWh / 1000.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *RaedianChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *RaedianChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
