package templates

import (
	"log"
	"math/rand"
	"time"

	"nems/internal/models"
)

type EaseeChargerPoller struct {
	Device models.Device
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "easee_charger",
			Name: "Easee Charger",
			Type: "cloud",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &EaseeChargerPoller{Device: device}
		},
	})
}

func (p *EaseeChargerPoller) Connect() error {
	log.Printf("EaseeChargerPoller: Authenticating with cloud using user %s", p.Device.Username)
	time.Sleep(500 * time.Millisecond)
	p.status = "online"
	return nil
}

func (p *EaseeChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *EaseeChargerPoller) Poll() (float64, float64, float64, error) {
	powerW := 0.0
	if rand.Float32() > 0.4 {
		powerW = 22000.0
	}
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, energyKwh, nil
}

func (p *EaseeChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *EaseeChargerPoller) Close() error {
	return nil
}
