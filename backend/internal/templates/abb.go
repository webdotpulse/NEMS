package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type AbbChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "abb_charger",
			Name:     "ABB Charger",
			Vendor:   "ABB",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &AbbChargerPoller{Device: device}
		},
	})
}

func (p *AbbChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] AbbChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] AbbChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] AbbChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *AbbChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *AbbChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(0x401C, 2, modbus.HOLDING_REGISTER)
	powerW := 0.0
	if err == nil {
		powerW = float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))
	}

	energyRegs, err := p.client.ReadRegisters(0x401E, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16 | uint32(energyRegs[1])) / 1000.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *AbbChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *AbbChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	curr := uint32(amps * 1000.0)
	highWord := uint16(curr >> 16)
	lowWord := uint16(curr & 0xFFFF)
	return p.client.WriteRegisters(0x4100, []uint16{highWord, lowWord})
}

func (p *AbbChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
