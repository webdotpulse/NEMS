package templates

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type HuaweiInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "huawei_inverter",
			Name: "Huawei Hybrid Inverter",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &HuaweiInverterPoller{Device: device}
		},
	})
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "huawei_dongle",
			Name: "Huawei Dongle Power Sensor",
			Type: "modbus",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &HuaweiDonglePoller{Device: device}
		},
	})
}

func (p *HuaweiInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("HuaweiInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("HuaweiInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *HuaweiInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *HuaweiInverterPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 1000.0 + rand.Float64()*3000.0
		batteryPowerW := -2000.0 + rand.Float64()*4000.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, energyKwh, nil
	}

	powerRegs, err := p.client.ReadRegisters(32080, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, err
	}
	powerW := float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))

	energyRegs, err := p.client.ReadRegisters(32106, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.01
	}

	batRegs, err := p.client.ReadRegisters(37001, 2, modbus.HOLDING_REGISTER)
	batteryPowerW := 0.0
	if err == nil {
		batteryPowerW = float64(int32(uint32(batRegs[0])<<16|uint32(batRegs[1]))) * -1.0
	}

	return powerW, batteryPowerW, energyKwh, nil
}

func (p *HuaweiInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *HuaweiInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

type HuaweiDonglePoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func (p *HuaweiDonglePoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("HuaweiDonglePoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("HuaweiDonglePoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("HuaweiDonglePoller: Connection failed, falling back to simulation mode (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *HuaweiDonglePoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *HuaweiDonglePoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := -2000.0 + rand.Float64()*4000.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, energyKwh, nil
	}

	powerRegs, err := p.client.ReadRegisters(37113, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, err
	}
	powerW := float64(int32(uint32(powerRegs[0])<<16|uint32(powerRegs[1]))) * -1.0

	energyRegs, err := p.client.ReadRegisters(37121, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.01
	}

	return powerW, 0, energyKwh, nil
}

func (p *HuaweiDonglePoller) GetDevice() models.Device {
	return p.Device
}

func (p *HuaweiDonglePoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
