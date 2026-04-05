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
	Device              models.Device
	client              *modbus.ModbusClient
	status              string
	lastPollTime        time.Time
	cachedPowerW        float64
	cachedBatteryPowerW float64
	cachedGridPowerW    float64
	cachedEnergyKwh     float64
	cachedSoc           float64
	cachedError         error
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "huawei_inverter",
			Name:     "Huawei Hybrid Inverter",
			Vendor:   "Huawei",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &HuaweiInverterPoller{Device: device}
		},
	})
}

func (p *HuaweiInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] HuaweiInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] HuaweiInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] HuaweiInverterPoller: Connection failed, falling back to simulation mode (%v)", err)
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
	if time.Since(p.lastPollTime) < 15*time.Second {
		return p.cachedPowerW, p.cachedBatteryPowerW, p.cachedGridPowerW, p.cachedEnergyKwh, p.cachedSoc, p.cachedError
	}

	p.lastPollTime = time.Now()

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

		p.cachedPowerW = powerW
		p.cachedBatteryPowerW = batteryPowerW
		p.cachedGridPowerW = gridPowerW
		p.cachedEnergyKwh = energyKwh
		p.cachedSoc = soc
		p.cachedError = nil

		return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
	}

	powerRegs, err := p.client.ReadRegisters(32080, 2, modbus.HOLDING_REGISTER)
	if err != nil {
		p.cachedError = err
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

	p.cachedPowerW = powerW
	p.cachedBatteryPowerW = batteryPowerW
	p.cachedGridPowerW = gridPowerW
	p.cachedEnergyKwh = energyKwh
	p.cachedSoc = soc
	p.cachedError = nil

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *HuaweiInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *HuaweiInverterPoller) DischargeBattery(powerW float64) error {
	log.Printf("[INFO] HuaweiInverterPoller: Commanding battery discharge at %.2f W (Simulated)", powerW)
	return nil
}

func (p *HuaweiInverterPoller) ChargeBattery(powerW float64) error {
	log.Printf("[INFO] HuaweiInverterPoller: Commanding battery charge at %.2f W (Simulated)", powerW)
	return nil
}

func (p *HuaweiInverterPoller) SetActivePowerLimit(powerW float64) error {
	if p.client == nil || p.status != "online" {
		log.Printf("[INFO] HuaweiInverterPoller: Cannot set active power limit, device offline")
		return nil
	}

	if powerW >= 100000.0 {
		// Release curtailment (No limit)
		err := p.client.WriteRegister(40118, 0)
		if err != nil {
			log.Printf("[ERROR] HuaweiInverterPoller: Failed to release active power limit (%v)", err)
			return err
		}
		log.Printf("[INFO] HuaweiInverterPoller: Released active power limit")
		return nil
	}

	// 40118 Active Power Control Mode -> 2
	err := p.client.WriteRegister(40118, 2)
	if err != nil {
		log.Printf("[ERROR] HuaweiInverterPoller: Failed to set Active Power Control Mode (%v)", err)
		return err
	}

	// 40120 Fixed active power derating. Unit: 0.1 kW
	val := uint32(powerW / 100.0)
	highWord := uint16(val >> 16)
	lowWord := uint16(val & 0xFFFF)
	err = p.client.WriteRegisters(40120, []uint16{highWord, lowWord})
	if err != nil {
		log.Printf("[ERROR] HuaweiInverterPoller: Failed to write Fixed active power derating (%v)", err)
		return err
	}

	log.Printf("[INFO] HuaweiInverterPoller: Setting active power limit to %.2f W", powerW)
	return nil
}

func (p *HuaweiInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
