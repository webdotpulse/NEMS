package templates

import (
	"bufio"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"nems/internal/models"

	"go.bug.st/serial"
)

type P1SerialPoller struct {
	Device models.Device
	powerW float64
	gridW  float64
	status string
	mu     sync.Mutex
	port   serial.Port
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "p1_serial",
			Name: "P1 Smart Meter (USB/Serial)",
			Type: "serial",
			Category: "meter",
		},
		NewPoller: func(d models.Device) models.DevicePoller {
			return &P1SerialPoller{Device: d, status: "offline"}
		},
	})
}

func (p *P1SerialPoller) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.port != nil {
		return nil
	}
	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	portName := p.Device.Host
	if portName == "" || portName == "localhost" {
		portName = "/dev/ttyUSB0"
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Printf("P1Serial [%s]: Failed to open serial port %s: %v", p.Device.Name, portName, err)
		p.status = "error"
		return err
	}
	p.port = port
	p.status = "online"
	return nil
}

func (p *P1SerialPoller) Poll() (float64, float64, float64, float64, float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.port == nil {
		if err := p.Connect(); err != nil {
			return 0, 0, 0, 0, 0, err
		}
	}

	reader := bufio.NewReader(p.port)
	p.port.SetReadTimeout(3 * time.Second)

	var importPower, exportPower float64
	var totalImport float64

	// Read a chunk of lines to find Telegram data
	for i := 0; i < 40; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		// 1.7.0 = Import Power (kW)
		if strings.HasPrefix(line, "1-0:1.7.0") {
			valStr := extractP1Value(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			importPower = val * 1000.0 // Convert to W
		}
		// 2.7.0 = Export Power (kW)
		if strings.HasPrefix(line, "1-0:2.7.0") {
			valStr := extractP1Value(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			exportPower = val * 1000.0 // Convert to W
		}
		// 1.8.1 / 1.8.2 = Total Import (kWh)
		if strings.HasPrefix(line, "1-0:1.8.1") || strings.HasPrefix(line, "1-0:1.8.2") {
			valStr := extractP1Value(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			totalImport += val // Keep in kWh
		}
	}

	p.gridW = importPower - exportPower
	p.powerW = 0
	p.status = "online"

	// Ensure we only read recent data next time
	p.port.ResetInputBuffer()

	return p.powerW, 0, p.gridW, totalImport, 0, nil
}

func extractP1Value(line string) string {
	start := strings.Index(line, "(")
	end := strings.Index(line, "*")
	if end == -1 {
		end = strings.Index(line, ")")
	}
	if start != -1 && end != -1 && end > start {
		return line[start+1 : end]
	}
	return "0"
}

func (p *P1SerialPoller) GetDevice() models.Device {
	return p.Device
}

func (p *P1SerialPoller) Status() string {
	return p.status
}

func (p *P1SerialPoller) Close() error {
	if p.port != nil {
		return p.port.Close()
	}
	return nil
}
