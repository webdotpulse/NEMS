package templates

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"nems/internal/models"

	"github.com/simonvetter/modbus"
)

type EnerlutionPoller struct {
	Device models.Device
	client *modbus.ModbusClient
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "enerlution_inverter",
			Name:     "Enerlution Hybrid Inverter",
			Vendor:   "Enerlution",
			Type:     "modbus",
			Category: "inverter",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &EnerlutionPoller{Device: device}
		},
	})
}

func (p *EnerlutionPoller) Connect() error {
	var addr string
	var conf *modbus.ClientConfiguration

	if strings.HasPrefix(p.Device.Host, "/dev/") || strings.HasPrefix(strings.ToUpper(p.Device.Host), "COM") {
		addr = "rtu://" + p.Device.Host
		conf = &modbus.ClientConfiguration{
			URL:      addr,
			Timeout:  3 * time.Second,
			Speed:    9600,
			DataBits: 8,
			Parity:   modbus.PARITY_NONE,
			StopBits: 1,
		}
	} else {
		addr = "tcp://" + p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
		conf = &modbus.ClientConfiguration{
			URL:     addr,
			Timeout: 3 * time.Second,
		}
	}

	log.Printf("EnerlutionPoller: Attempting Modbus connection to %s (ID: %d)", addr, p.Device.ModbusID)

	client, err := modbus.NewClient(conf)
	if err != nil {
		log.Printf("EnerlutionPoller: Client setup failed (%v)", err)
		p.status = "error"
		return err
	}
	client.SetUnitId(uint8(p.Device.ModbusID))

	if err := client.Open(); err != nil {
		log.Printf("EnerlutionPoller: Connection failed (%v)", err)
		p.status = "error"
		return err
	}

	p.client = client
	p.status = "online"
	return nil
}

func (p *EnerlutionPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *EnerlutionPoller) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

func (p *EnerlutionPoller) GetDevice() models.Device {
	return p.Device
}

// decodeS32 parses two Modbus registers into a signed 32-bit integer (S32).
func decodeS32(regs []uint16) int32 {
	if len(regs) < 2 {
		return 0
	}
	return int32(uint32(regs[0])<<16 | uint32(regs[1]))
}

// decodeFloat32 parses two Modbus registers into an IEEE 754 32-bit float.
func decodeFloat32(regs []uint16) float32 {
	if len(regs) < 2 {
		return 0
	}
	bits := uint32(regs[0])<<16 | uint32(regs[1])
	return math.Float32frombits(bits)
}

// decodeS16 parses one Modbus register into a signed 16-bit integer (S16).
func decodeS16(reg uint16) int16 {
	return int16(reg)
}

func (p *EnerlutionPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.client == nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("device offline")
	}

	// First, fetch telemetry which populates everything for monitoring.
	t, err := p.FetchTelemetry()
	if err != nil {
		p.status = "error"
		return 0, 0, 0, 0, 0, err
	}

	p.status = "online" // Reset error status if successful

	// The Telemetry struct holds our parsed values, map them to the generic interface.
	powerW := 0.0
	batteryPowerW := 0.0
	soc := 0.0
	gridPowerW := 0.0
	energyKwh := 0.0

	if t != nil {
		powerW = t.OutputActivePowerW

		if p.Device.HasBattery {
			batteryPowerW = t.BatteryPowerW
			soc = t.BatterySOC
		}

		if p.Device.HasGridMeter {
			gridPowerW = t.ThirdPartyMeterPowerW
		}

		energyKwh = t.TotalEnergyYieldKwh
	}

	return powerW, batteryPowerW, gridPowerW, energyKwh, soc, nil
}

// SetActivePowerLimit implements the models.InverterController interface
func (p *EnerlutionPoller) SetActivePowerLimit(powerW float64) error {
	if p.client == nil || p.status != "online" {
		err := fmt.Errorf("device offline")
		log.Printf("EnerlutionPoller: Cannot set active power limit, device offline")
		return err
	}

	// Active Power Control Mode (40400) -> 1 (Fixed active power)
	err := p.client.WriteRegister(40400, 1)
	if err != nil {
		log.Printf("EnerlutionPoller: Failed to set Active Power Control Mode (%v)", err)
		return err
	}

	// Fixed Active Power Value (40441) -> Length 4 (Wait, the instructions say:
	// 40441 (Hex 9DF9): Fixed Active Power Value (Length 4, S32). Unit: W.
	// Oh, length 4 bytes means 2 registers for S32.

	powerInt := int32(powerW)
	highWord := uint16(uint32(powerInt) >> 16)
	lowWord := uint16(uint32(powerInt) & 0xFFFF)

	err = p.client.WriteRegisters(40441, []uint16{highWord, lowWord})
	if err != nil {
		log.Printf("EnerlutionPoller: Failed to write Fixed Active Power Value (%v)", err)
		return err
	}

	log.Printf("EnerlutionPoller: Successfully set active power limit to %.2f W", powerW)
	return nil
}

// decodeASCII parses an array of Modbus registers into an ASCII string
func decodeASCII(regs []uint16) string {
	var sb strings.Builder
	for _, reg := range regs {
		b1 := byte(reg >> 8)
		b2 := byte(reg & 0xFF)
		if b1 != 0 {
			sb.WriteByte(b1)
		}
		if b2 != 0 {
			sb.WriteByte(b2)
		}
	}
	return strings.TrimSpace(sb.String())
}

type EnerlutionTelemetry struct {
	// Device Information
	Manufacturer string
	Model        string
	SerialNumber string

	// AC / Grid Statistics
	InverterState    uint16
	OutputActivePowerW float64
	GridRVoltageV    float64
	GridRCurrentA    float64
	GridFrequencyHz  float64
	TotalEnergyYieldKwh float64

	// PV (DC)
	PV1VoltageV     float64
	PV1CurrentA     float64
	PV2VoltageV     float64
	PV2CurrentA     float64
	TotalPVPowerW   float64
	TotalPVEnergyKwh float64

	// Battery
	BatteryState       uint16
	BatteryPowerW      float64
	BatteryVoltageV    float64
	BatteryCurrentA    float64
	BatterySOC         float64
	BatterySOH         float64

	// Meter
	ThirdPartyMeterPowerW float64
	ThirdPartyMeterEnergyKwh float64
	ImportEnergyKwh       float64
	ExportEnergyKwh       float64

	// Writeable Settings & Controls
	EMSMode                   uint16
	ActivePowerControlMode    uint16
	FixedActivePowerPercent   uint16
	FixedActivePowerValueW    float64
}

// isNetworkErr returns true if the error is related to connection dropping
func isNetworkErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "EOF") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "broken pipe") {
		return true
	}
	// A Modbus protocol error (e.g. exception response) is safe to skip
	if strings.HasPrefix(msg, "modbus:") {
		return false
	}
	// Assume any other unknown error is a network/system error
	return true
}

func (p *EnerlutionPoller) FetchTelemetry() (*EnerlutionTelemetry, error) {
	if p.client == nil || p.status != "online" {
		return nil, fmt.Errorf("device offline")
	}

	t := &EnerlutionTelemetry{}

	// Read a block of registers, aborting the entire fetch on network failures (timeout/EOF).
	// Modbus logic errors (like illegal data address) are ignored to handle skipped/unsupported registers.
	read := func(addr uint16, length uint16) ([]uint16, error) {
		regs, err := p.client.ReadRegisters(addr, length, modbus.HOLDING_REGISTER)
		if isNetworkErr(err) {
			return nil, err
		}
		if err != nil {
			return nil, nil // Ignored
		}
		return regs, nil
	}

	var regs []uint16
	var err error

	// A. Device Information
	if regs, err = read(30010, 8); err != nil { return nil, err }
	if len(regs) == 8 { t.Manufacturer = decodeASCII(regs) }

	if regs, err = read(30018, 8); err != nil { return nil, err }
	if len(regs) == 8 { t.Model = decodeASCII(regs) }

	if regs, err = read(30026, 8); err != nil { return nil, err }
	if len(regs) == 8 { t.SerialNumber = decodeASCII(regs) }

	// B. Inverter & Grid Statistics (AC)
	if regs, err = read(30115, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.InverterState = regs[0] }

	if regs, err = read(30100, 2); err != nil { return nil, err }
	if len(regs) == 2 { t.OutputActivePowerW = float64(decodeS32(regs)) }

	if regs, err = read(30131, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.GridRVoltageV = float64(regs[0]) * 0.1 }

	if regs, err = read(30132, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.GridRCurrentA = float64(regs[0]) * 0.1 }

	if regs, err = read(30140, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.GridFrequencyHz = float64(regs[0]) * 0.01 }

	if regs, err = read(30154, 2); err != nil { return nil, err }
	if len(regs) == 2 { t.TotalEnergyYieldKwh = float64(decodeFloat32(regs)) }

	// C. PV / Solar Panels (DC)
	if regs, err = read(30119, 4); err != nil { return nil, err }
	if len(regs) == 4 {
		t.PV1VoltageV = float64(regs[0]) * 0.1
		t.PV1CurrentA = float64(regs[1]) * 0.1
		t.PV2VoltageV = float64(regs[2]) * 0.1
		t.PV2CurrentA = float64(regs[3]) * 0.1
	}

	if regs, err = read(30127, 2); err != nil { return nil, err }
	if len(regs) == 2 { t.TotalPVPowerW = float64(decodeS32(regs)) }

	if regs, err = read(30129, 2); err != nil { return nil, err }
	if len(regs) == 2 { t.TotalPVEnergyKwh = float64(decodeFloat32(regs)) }

	// D. Battery Integration
	if p.Device.HasBattery {
		if regs, err = read(30161, 1); err != nil { return nil, err }
		if len(regs) == 1 { t.BatteryState = regs[0] }

		if regs, err = read(30162, 2); err != nil { return nil, err }
		if len(regs) == 2 { t.BatteryPowerW = float64(decodeS32(regs)) }

		if regs, err = read(30164, 2); err != nil { return nil, err }
		if len(regs) == 2 {
			t.BatteryVoltageV = float64(regs[0]) * 0.1
			t.BatteryCurrentA = float64(decodeS16(regs[1])) * 0.1
		}

		if regs, err = read(30182, 1); err != nil { return nil, err }
		if len(regs) == 1 { t.BatterySOC = float64(regs[0]) }

		if regs, err = read(30249, 1); err != nil { return nil, err }
		if len(regs) == 1 { t.BatterySOH = float64(regs[0]) }
	}

	// E. Grid Metering
	if p.Device.HasGridMeter {
		if regs, err = read(30110, 2); err != nil { return nil, err }
		if len(regs) == 2 { t.ThirdPartyMeterPowerW = float64(decodeS32(regs)) }

		if regs, err = read(30112, 2); err != nil { return nil, err }
		if len(regs) == 2 { t.ThirdPartyMeterEnergyKwh = float64(decodeFloat32(regs)) }

		if regs, err = read(30156, 4); err != nil { return nil, err }
		if len(regs) == 4 {
			t.ImportEnergyKwh = float64(decodeFloat32(regs[0:2]))
			t.ExportEnergyKwh = float64(decodeFloat32(regs[2:4]))
		}
	}

	// F. Writeable Settings & Controls
	if regs, err = read(40907, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.EMSMode = regs[0] }

	if regs, err = read(40400, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.ActivePowerControlMode = regs[0] }

	if regs, err = read(40410, 1); err != nil { return nil, err }
	if len(regs) == 1 { t.FixedActivePowerPercent = regs[0] }

	if regs, err = read(40441, 2); err != nil { return nil, err }
	if len(regs) == 2 { t.FixedActivePowerValueW = float64(decodeS32(regs)) }

	return t, nil
}
