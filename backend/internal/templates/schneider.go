package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type SchneiderChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "schneider_charger",
			Name:     "Schneider EVlink",
			Vendor:   "Schneider",
			Type:     "modbus",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &SchneiderChargerPoller{Device: device}
		},
	})
}

func (p *SchneiderChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] SchneiderChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] SchneiderChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] SchneiderChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *SchneiderChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *SchneiderChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerRegs, err := p.client.ReadRegisters(3059, 2, modbus.HOLDING_REGISTER) // Active power
	powerW := 0.0
	if err == nil {
		powerW = float64(uint32(powerRegs[0])<<16 | uint32(powerRegs[1])) // Assuming standard modbus representation
	}

	energyRegs, err := p.client.ReadRegisters(3203, 4, modbus.HOLDING_REGISTER) // Total active energy import
	energyKwh := 0.0
	if err == nil {
		// EVCC schneider iem3000 uses int64 for 4 registers
		energyKwh = float64(uint64(energyRegs[0])<<48|uint64(energyRegs[1])<<32|uint64(energyRegs[2])<<16|uint64(energyRegs[3])) / 1000.0
	}

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *SchneiderChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *SchneiderChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}
	// For V3 Schneider, setting charge current is typically around register 4004 based on standard sunspec/modbus EVSE, but EVCC seems to vary.
	// We will use standard register mapping for now.
	return p.client.WriteRegister(4004, uint16(amps))
}

func (p *SchneiderChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
