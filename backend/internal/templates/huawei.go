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

func (p *HuaweiInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		powerW := 1000.0 + rand.Float64()*3000.0
		batteryPowerW := 0.0
		soc := 0.0
		if p.Device.HasBattery {
			batteryPowerW = -2000.0 + rand.Float64()*4000.0
			soc = 20.0 + rand.Float64()*60.0
		}
		gridPowerW := 0.0
		if p.Device.HasGridMeter {
			gridPowerW = -2000.0 + rand.Float64()*4000.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
	}

	powerRegs, err := p.client.ReadRegisters(32080, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	powerW := float64(int32(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])))

	energyRegs, err := p.client.ReadRegisters(32106, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) * 0.01
	}

	batteryPowerW := 0.0
	soc := 0.0
	if p.Device.HasBattery {
		batRegs, err := p.client.ReadRegisters(37001, 2, modbus.HOLDING_REGISTER)
		if err == nil {
			batteryPowerW = float64(int32(uint32(batRegs[0])<<16|uint32(batRegs[1]))) * -1.0
		}
		socRegs, err := p.client.ReadRegisters(37004, 1, modbus.HOLDING_REGISTER)
		if err == nil {
			soc = float64(socRegs[0]) / 10.0
		}
	}

	gridPowerW := 0.0
	if p.Device.HasGridMeter {
		gridRegs, err := p.client.ReadRegisters(37113, 2, modbus.HOLDING_REGISTER)
		if err == nil {
			gridPowerW = float64(int32(uint32(gridRegs[0])<<16|uint32(gridRegs[1]))) * -1.0
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
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
