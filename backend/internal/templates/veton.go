package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type VetonChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "veton_charger",
			Name:     "Veton / Phoenix CharX",
			Vendor:   "Veton",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &VetonChargerPoller{Device: device}
		},
	})

	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "blitzpower_charger",
			Name:     "Blitzpower / Phoenix CharX",
			Vendor:   "Blitzpower",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &VetonChargerPoller{Device: device}
		},
	})
}

func (p *VetonChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] VetonChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] VetonChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] VetonChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *VetonChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *VetonChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	// Phoenix Charx 1000 offset + reg
	// 244 (Power, mW)
	powerRegs, err := p.client.ReadRegisters(1244, 2, modbus.HOLDING_REGISTER)
	powerW := 0.0
	if err == nil {
		powerW = float64(int32(uint32(powerRegs[0])<<16|uint32(powerRegs[1]))) / 1000.0
	}

	// 250 (Energy, Wh)
	energyRegs, err := p.client.ReadRegisters(1250, 4, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(int64(uint64(energyRegs[0])<<48|uint64(energyRegs[1])<<32|uint64(energyRegs[2])<<16|uint64(energyRegs[3]))) / 1000.0 / 1000.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *VetonChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *VetonChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	return p.client.WriteRegister(1301, uint16(amps))
}

func (p *VetonChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
