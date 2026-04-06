package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type KebaChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "keba_charger",
			Name:     "KEBA KeContact P30/P40",
			Vendor:   "KEBA",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &KebaChargerPoller{Device: device}
		},
	})
}

func (p *KebaChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] KebaChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] KebaChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] KebaChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *KebaChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *KebaChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(1020, 2, modbus.HOLDING_REGISTER) // Total active power (mW)
	powerW := 0.0
	if err == nil {
		powerW = float64(uint32(powerRegs[0])<<16|uint32(powerRegs[1])) / 1000.0
	}

	energyRegs, err := p.client.ReadRegisters(1036, 2, modbus.HOLDING_REGISTER) // Total Energy (10Wh)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) / 100.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *KebaChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *KebaChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	curr := uint32(amps * 1000.0) // mA
	highWord := uint16(curr >> 16)
	lowWord := uint16(curr & 0xFFFF)
	return p.client.WriteRegisters(5004, []uint16{highWord, lowWord}) // Set current limit
}

func (p *KebaChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
