package templates

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"nems/internal/models"
)

type P1NetworkPoller struct {
	Device models.Device
	powerW float64
	gridW  float64
	status string
	mu     sync.Mutex
	conn   net.Conn
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "p1_network",
			Name:     "P1 Smart Meter (Network/TCP)",
			Vendor:   "Generic",
			Type:     "network",
			Category: "meter",
		},
		NewPoller: func(d models.Device) models.DevicePoller {
			return &P1NetworkPoller{Device: d, status: "offline"}
		},
	})
}

func (p *P1NetworkPoller) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		return nil
	}
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Printf("[ERROR] P1Network [%s]: Failed to connect to %s: %v", p.Device.Name, addr, err)
		p.status = "error"
		return err
	}
	p.conn = conn
	p.status = "online"
	return nil
}

func (p *P1NetworkPoller) Poll() (float64, float64, float64, float64, float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil {
		if err := p.Connect(); err != nil {
			return 0, 0, 0, 0, 0, err
		}
	}

	p.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	reader := bufio.NewReader(p.conn)

	var importPower, exportPower float64
	var totalImport float64

	for i := 0; i < 40; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			p.conn.Close()
			p.conn = nil
			break
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "1-0:1.7.0") {
			valStr := extractP1ValueNetwork(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			importPower = val * 1000.0
		}
		if strings.HasPrefix(line, "1-0:2.7.0") {
			valStr := extractP1ValueNetwork(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			exportPower = val * 1000.0
		}
		if strings.HasPrefix(line, "1-0:1.8.1") || strings.HasPrefix(line, "1-0:1.8.2") {
			valStr := extractP1ValueNetwork(line)
			val, _ := strconv.ParseFloat(valStr, 64)
			totalImport += val
		}
	}

	p.gridW = importPower - exportPower
	p.powerW = 0
	p.status = "online"

	return p.powerW, 0, p.gridW, totalImport, 0, nil
}

func extractP1ValueNetwork(line string) string {
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

func (p *P1NetworkPoller) GetDevice() models.Device {
	return p.Device
}

func (p *P1NetworkPoller) Status() string {
	return p.status
}

func (p *P1NetworkPoller) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
