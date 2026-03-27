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

type HomeWizardMeterPoller struct {
	Device models.Device
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "homewizard_meter",
			Name: "HomeWizard Meter",
			Type: "rest",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &HomeWizardMeterPoller{Device: device}
		},
	})
}

func (p *HomeWizardMeterPoller) Connect() error {
	addr := p.Device.Host
	if p.Device.Port != 80 && p.Device.Port != 0 {
		addr = addr + ":" + strconv.Itoa(p.Device.Port)
	}
	log.Printf("HomeWizardMeterPoller: Attempting local REST API connection to http://%s/api/v1/data", addr)

	url := fmt.Sprintf("http://%s/api/v1/data", addr)
	client := http.Client{Timeout: 2 * time.Second}
	_, err := client.Get(url)
	if err != nil {
		log.Printf("HomeWizardMeterPoller: Connection failed, falling back to simulation (%v)", err)
		p.status = "error"
		return nil
	}

	p.status = "online"
	return nil
}

func (p *HomeWizardMeterPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *HomeWizardMeterPoller) Poll() (float64, float64, float64, error) {
	if p.status != "online" {
		powerW := -1500.0 + rand.Float64()*3500.0
		energyKwh := powerW * (5.0 / 3600.0) / 1000.0
		return powerW, 0, energyKwh, nil
	}

	addr := p.Device.Host
	if p.Device.Port != 80 && p.Device.Port != 0 {
		addr = addr + ":" + strconv.Itoa(p.Device.Port)
	}
	url := fmt.Sprintf("http://%s/api/v1/data", addr)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	var data struct {
		ActivePowerW         float64 `json:"active_power_w"`
		TotalEnergyImportKwh float64 `json:"total_power_import_kwh"`
		TotalEnergyExportKwh float64 `json:"total_power_export_kwh"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, 0, err
	}

	return data.ActivePowerW, 0, data.TotalEnergyImportKwh - data.TotalEnergyExportKwh, nil
}

func (p *HomeWizardMeterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *HomeWizardMeterPoller) Close() error {
	return nil
}
