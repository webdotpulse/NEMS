package templates

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SmaInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "sma_inverter",
			Name: "SMA Inverter",
			Type: "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SmaInverterPoller{Device: device}
		},
	})
}

func (p *SmaInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("SmaInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("SmaInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("SmaInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SmaInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SmaInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 1500.0 + rand.Float64()*2500.0
		batteryPowerW := -500.0 + rand.Float64()*1500.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, 0, energyKwh, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(30775, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	powerW := float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))

	energyRegs, err := p.client.ReadRegisters(30529, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyWh := float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1]))
		energyKwh = energyWh / 1000.0
	}

	batRegs, err := p.client.ReadRegisters(31393, 2, modbus.HOLDING_REGISTER)
	batteryPowerW := 0.0
	if err == nil {
		batteryPowerW = float64(int32(uint32(batRegs[0])<<16 | uint32(batRegs[1])))
	}

	return powerW, batteryPowerW, 0, energyKwh, 0, nil
}

func (p *SmaInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SmaInverterPoller) SetActivePowerLimit(powerW float64) error {
	log.Printf("SmaInverterPoller: Setting active power limit to %.2f W (Simulated)", powerW)
	return nil
}

func (p *SmaInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
