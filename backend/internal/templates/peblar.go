package templates

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"nems/internal/models"
)

type PeblarChargerPoller struct {
	Device models.Device
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "peblar_charger",
			Name: "Peblar Charger",
			Type: "rest",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &PeblarChargerPoller{Device: device}
		},
	})
}

func (p *PeblarChargerPoller) Connect() error {
	addr := p.Device.Host + ":" + strconv.Itoa(p.Device.Port)
	log.Printf("PeblarChargerPoller: Attempting REST API connection to %s", addr)

	url := fmt.Sprintf("http://%s/api/v1/system", addr)
	client := http.Client{Timeout: 2 * time.Second}
	_, err := client.Get(url)
	if err != nil {
		log.Printf("PeblarChargerPoller: Connection failed, falling back to simulation (%v)", err)
		p.status = "error"
		return nil
	}

	p.status = "online"
	return nil
}

func (p *PeblarChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *PeblarChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" {
		powerW := 0.0
		if rand.Float32() > 0.5 {
			powerW = 11000.0
		}
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, 0, energyKwh, 0, nil
	}

	url := fmt.Sprintf("http://%s:%d/api/v1/meter", p.Device.Host, p.Device.Port)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	var data struct {
		Power  float64 `json:"power"`
		Energy float64 `json:"energy"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return data.Power, 0, 0, data.Energy, 0, nil
}

func (p *PeblarChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *PeblarChargerPoller) SetChargeCurrent(amps float64) error {
	log.Printf("PeblarChargerPoller: Setting charge current to %.2f A (Simulated)", amps)
	return nil
}

func (p *PeblarChargerPoller) Close() error {
	return nil
}
