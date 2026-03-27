package templates

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type GenericModbusChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
	Prefix string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "alfen_charger",
			Name: "Alfen Charger",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GenericModbusChargerPoller{Device: device, Prefix: "AlfenChargerPoller"}
		},
	})
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "bender_charger",
			Name: "Bender Charger",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GenericModbusChargerPoller{Device: device, Prefix: "BenderChargerPoller"}
		},
	})
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "phoenix_charger",
			Name: "Phoenix Contact Charx Charger",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GenericModbusChargerPoller{Device: device, Prefix: "PhoenixChargerPoller"}
		},
	})
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

func (p *GenericModbusChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 0.0
		if rand.Float32() > 0.6 {
			powerW = 7400.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, 0, energyKwh, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(344, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	powerW := float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *GenericModbusChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *GenericModbusChargerPoller) SetChargeCurrent(amps float64) error {
	log.Printf("%s: Setting charge current to %.2f A (Simulated)", p.Prefix, amps)
	return nil
}

func (p *GenericModbusChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
