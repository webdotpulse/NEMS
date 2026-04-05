package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nems/internal/models"
)

type GenericThermostatPoller struct {
	Device models.Device
	client *http.Client
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "generic_thermostat",
			Name: "Generic HTTP Thermostat",
			Vendor: "Generic",
			Type: "rest",
			Category: "thermostat",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GenericThermostatPoller{
				Device: device,
				client: &http.Client{Timeout: 5 * time.Second},
			}
		},
	})
}

func (p *GenericThermostatPoller) Connect() error {
	return nil
}

func (p *GenericThermostatPoller) Status() string {
	return "online"
}

func (p *GenericThermostatPoller) Poll() (float64, float64, float64, float64, float64, error) {
	// A generic thermostat doesn't inherently report power usage unless extended.
	// For now, we return 0 for metrics.
	return 0, 0, 0, 0, 0, nil
}

func (p *GenericThermostatPoller) GetDevice() models.Device {
	return p.Device
}

func (p *GenericThermostatPoller) Close() error {
	return nil
}

func (p *GenericThermostatPoller) SetTargetTemperature(temp float64) error {
	payload := map[string]interface{}{
		"target_temperature": temp,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%d/temperature", p.Device.Host, p.Device.Port)
	if strings.HasPrefix(p.Device.Host, "http") {
		url = fmt.Sprintf("%s:%d/temperature", p.Device.Host, p.Device.Port)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if p.Device.Username != "" || p.Device.Password != "" {
		req.SetBasicAuth(p.Device.Username, p.Device.Password)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) // consume body

	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
