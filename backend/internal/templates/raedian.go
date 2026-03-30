package templates

import (
	"log"
	"strconv"
	"strings"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type RaedianChargerPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string

	// Device Information
	SerialNumber    uint64
	FirmwareVersion uint32
	MaxRatedCurrent float64 // in A

	// Telemetry
	SocketLockState      uint32
	ChargingState        uint32
	CurrentChargingLimit float64 // in A
	CurrentPhase1        float64 // in A
	CurrentPhase2        float64 // in A
	CurrentPhase3        float64 // in A
	VoltagePhase1        float64 // in V
	VoltagePhase2        float64 // in V
	VoltagePhase3        float64 // in V
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "raedian_charger",
			Name:     "Raedian NEX/NEO AC Wallbox",
			Vendor:   "Raedian",
			Type:     "modbus",
			Category: "charger",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &RaedianChargerPoller{Device: device}
		},
	})
}

func (p *RaedianChargerPoller) Connect() error {
	addr := "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("[INFO] RaedianChargerPoller: Attempting Modbus TCP connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("[ERROR] RaedianChargerPoller: Client setup failed (%v)", err)
		p.status = "error"
		return nil
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("[ERROR] RaedianChargerPoller: Connection failed (%v)", err)
		p.status = "error"
		return nil
	}

	p.client = client
	p.status = "online"

	// Read Device Information (Once on connect)
	// 0x8000 (32768): Serial Number (Size: 4)
	// 0x8004 (32772): Firmware Version (Size: 2)
	// 0x8006 (32774): Max Rated/Settable Current (Size: 2) -> Scale 0.001A

	// Read from 32768 to 32775 (8 registers total)
	devRegs, err := p.client.ReadRegisters(32768, 8, modbus.HOLDING_REGISTER)
	if err == nil {
		p.SerialNumber = uint64(devRegs[0])<<48 | uint64(devRegs[1])<<32 | uint64(devRegs[2])<<16 | uint64(devRegs[3])
		p.FirmwareVersion = uint32(devRegs[4])<<16 | uint32(devRegs[5])
		maxCurrentRaw := uint32(devRegs[6])<<16 | uint32(devRegs[7])
		p.MaxRatedCurrent = float64(maxCurrentRaw) * 0.001
	} else {
		log.Printf("[ERROR] RaedianChargerPoller: Failed to read device info: %v", err)
	}

	return nil
}

func (p *RaedianChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *RaedianChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *RaedianChargerPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

func (p *RaedianChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, nil
	}

	// Read from 0x800A (32778) to 0x801F (32799)
	// That's 22 registers total (32799 - 32778 + 1 = 22)
	// Registers we need:
	// 32778: Socket Lock State (2)
	// 32780: Charging State (2)
	// 32782: Charging Current Limit (2)
	// 32784: Phase 1 Current (2)
	// 32786: Phase 2 Current (2)
	// 32788: Phase 3 Current (2)
	// 32790: Phase 1 Voltage (2)
	// 32792: Phase 2 Voltage (2)
	// 32794: Phase 3 Voltage (2)
	// 32796: Active Power (2)
	// 32798: Energy Delivered in Session (2)
	regs, err := p.client.ReadRegisters(32778, 22, modbus.HOLDING_REGISTER)
	if err != nil {
		log.Printf("[ERROR] RaedianChargerPoller: Failed to read telemetry: %v", err)
		if strings.HasPrefix(err.Error(), "modbus:") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "connection reset") {
			p.status = "error"
			p.client.Close()
		}
		return 0, 0, 0, 0, 0, err
	}

	p.SocketLockState = uint32(regs[0])<<16 | uint32(regs[1])
	p.ChargingState = uint32(regs[2])<<16 | uint32(regs[3])
	p.CurrentChargingLimit = float64(uint32(regs[4])<<16 | uint32(regs[5])) * 0.001
	p.CurrentPhase1 = float64(uint32(regs[6])<<16 | uint32(regs[7])) * 0.001
	p.CurrentPhase2 = float64(uint32(regs[8])<<16 | uint32(regs[9])) * 0.001
	p.CurrentPhase3 = float64(uint32(regs[10])<<16 | uint32(regs[11])) * 0.001
	p.VoltagePhase1 = float64(uint32(regs[12])<<16 | uint32(regs[13])) * 0.1
	p.VoltagePhase2 = float64(uint32(regs[14])<<16 | uint32(regs[15])) * 0.1
	p.VoltagePhase3 = float64(uint32(regs[16])<<16 | uint32(regs[17])) * 0.1

	// Active Power is at offset 18 (32796 - 32778 = 18)
	// Energy Delivered is at offset 20 (32798 - 32778 = 20)
	powerW := float64(uint32(regs[18])<<16 | uint32(regs[19]))

	energyWh := float64(uint32(regs[20])<<16 | uint32(regs[21]))
	energyKwh := energyWh / 1000.0

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *RaedianChargerPoller) StartCharging() error {
	if p.client == nil {
		return nil
	}
	// Write 0x00 to 33029 (Start/Stop Charging Session)
	return p.client.WriteRegister(33029, 0x00)
}

func (p *RaedianChargerPoller) StopCharging() error {
	if p.client == nil {
		return nil
	}
	// Write 0x01 to 33029 (Start/Stop Charging Session)
	return p.client.WriteRegister(33029, 0x01)
}

func (p *RaedianChargerPoller) SetChargeCurrent(amps float64) error {
	if p.client == nil {
		return nil
	}

	// Mode 3 valid range is 6A to 32A.
	// Setting a limit between 0 and 5999 pauses the charging session.
	// Scale: 0.001A (amps * 1000)

	val := uint32(amps * 1000)
	// Write to 0x8100 (33024) (Size: 2)

	// Since Size is 2 (32-bit value), we need to write multiple registers
	err := p.client.WriteRegisters(33024, []uint16{uint16(val >> 16), uint16(val & 0xFFFF)})
	if err != nil {
		log.Printf("[ERROR] RaedianChargerPoller: Failed to set charge current %.2f: %v", amps, err)
		return err
	}

	return nil
}

func (p *RaedianChargerPoller) SetChargingPhase(phases int) error {
	if p.client == nil {
		return nil
	}

	// 0x8102 (33026): Set Charging Phase (Size: 1)
	// 0x01 = Single Phase
	// 0x02 = Three Phases

	val := uint16(0x01)
	if phases == 3 {
		val = 0x02
	}

	return p.client.WriteRegister(33026, val)
}
