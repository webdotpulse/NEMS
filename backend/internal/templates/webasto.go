package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type WebastoChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "webasto_charger",
			Name:     "Webasto Next Charger",
			Vendor:   "Webasto",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &WebastoChargerPoller{Device: device}
		},
	})
}

func (p *WebastoChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] WebastoChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] WebastoChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] WebastoChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"

	// Send life bit
	p.client.WriteRegister(6000, 1)
	return nil
}

func (p *WebastoChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *WebastoChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	p.client.WriteRegister(6000, 1) // Heartbeat

	powerRegs, err := p.client.ReadRegisters(1020, 2, modbus.HOLDING_REGISTER)
	powerW := 0.0
	if err == nil {
		powerW = float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1]))
	}

	energyRegs, err := p.client.ReadRegisters(1036, 2, modbus.HOLDING_REGISTER)
	energyKwh := 0.0
	if err == nil {
		energyKwh = float64(uint32(energyRegs[0])<<16|uint32(energyRegs[1])) / 1000.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *WebastoChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *WebastoChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	return p.client.WriteRegister(5004, uint16(amps))
}

func (p *WebastoChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
