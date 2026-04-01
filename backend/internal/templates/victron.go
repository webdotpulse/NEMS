package templates

import (
	"log"
	"strconv"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type VictronInverterPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "victron_inverter",
			Name:     "Victron Inverter/GX",
			Vendor:   "Victron",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &VictronInverterPoller{Device: device}
		},
	})
}

func (p *VictronInverterPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] VictronInverterPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] VictronInverterPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	// Victron often uses ID 100 for system/GX, but we respect Device.ModbusID
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] VictronInverterPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *VictronInverterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *VictronInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW := 0.0
	batteryPowerW := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0
	soc := 0.0

	// CCGX System ID (often 100)
	sysID := uint8(100)
	if p.Device.ModbusID > 0 {
		sysID = uint8(p.Device.ModbusID)
	}
	p.client.SetUnitId(sysID)

	// PV Power (850 or 808 - total PV power)
	powerRegs, err := p.client.ReadRegisters(850, 1, modbus.HOLDING_REGISTER) // W
	if err == nil {
		powerW = float64(powerRegs[0])
	}

	if p.Device.HasBattery {
		// Battery Power (842)
		batRegs, err := p.client.ReadRegisters(842, 1, modbus.HOLDING_REGISTER) // W
		if err == nil {
			batteryPowerW = float64(int16(batRegs[0]))
		}
		// SOC (843)
		socRegs, err := p.client.ReadRegisters(843, 1, modbus.HOLDING_REGISTER) // %
		if err == nil {
			soc = float64(socRegs[0])
		}
	}

	if p.Device.HasGridMeter {
		// Grid Power (820)
		gridRegs, err := p.client.ReadRegisters(820, 1, modbus.HOLDING_REGISTER) // W
		if err == nil {
			gridPowerW = float64(int16(gridRegs[0]))
		}
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

func (p *VictronInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *VictronInverterPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
