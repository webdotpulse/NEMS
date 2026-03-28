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

type GenericRelayPoller struct {
	Device models.Device
	client *http.Client
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "generic_relay",
			Name: "Generic HTTP Relay",
			Vendor: "Generic",
			Type: "rest",
			Category: "relay",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &GenericRelayPoller{
				Device: device,
				client: &http.Client{Timeout: 5 * time.Second},
			}
		},
	})
}

func (p *GenericRelayPoller) Connect() error {
	return nil
}

func (p *GenericRelayPoller) Status() string {
	return "online"
}

func (p *GenericRelayPoller) Poll() (float64, float64, float64, float64, float64, error) {
	// A generic relay doesn't inherently report power usage unless extended.
	// For now, we return 0 for metrics.
	return 0, 0, 0, 0, 0, nil
}

func (p *GenericRelayPoller) GetDevice() models.Device {
	return p.Device
}

func (p *GenericRelayPoller) Close() error {
	return nil
}

func (p *GenericRelayPoller) SetState(on bool) error {
	state := "off"
	if on {
		state = "on"
	}

	payload := map[string]interface{}{
		"state": state,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%d/state", p.Device.Host, p.Device.Port)
	if strings.HasPrefix(p.Device.Host, "http") {
		url = fmt.Sprintf("%s:%d/state", p.Device.Host, p.Device.Port)
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
