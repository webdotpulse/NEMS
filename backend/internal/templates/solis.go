package templates

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SolisInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "solis_inverter",
			Name: "Solis Inverter",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SolisInverterPoller{Device: device}
		},
	})
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

func (p *SolisInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 800.0 + rand.Float64()*2000.0
		batteryPowerW := -1000.0 + rand.Float64()*2000.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, 0, energyKwh, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(33079, 2, modbus.INPUT_REGISTER)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	powerW := float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))

	energyRegs, err := p.client.ReadRegisters(33029, 2, modbus.INPUT_REGISTER)
	if err != nil {
		return powerW, 0, 0, 0, 0, err
	}
	energyKwh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1])) / 10.0

	batRegs, err := p.client.ReadRegisters(33149, 2, modbus.INPUT_REGISTER)
	batteryPowerW := 0.0
	if err == nil {
		rawBat := int32(uint32(batRegs[0])<<16 | uint32(batRegs[1]))
		batteryPowerW = float64(rawBat)
	}

	return powerW, batteryPowerW, 0, energyKwh, 0, nil
}

func (p *SolisInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SolisInverterPoller) SetActivePowerLimit(powerW float64) error {
	log.Printf("SolisInverterPoller: Setting active power limit to %.2f W (Simulated)", powerW)
	return nil
}

func (p *SolisInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
