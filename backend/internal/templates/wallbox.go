package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nems/internal/models"
)

type WallboxChargerPoller struct {
	Device     models.Device
	status     string
	token      string
	chargerID  int
	httpClient *http.Client
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "wallbox_charger",
			Name:     "Wallbox Pulsar/Commander",
			Vendor:   "Wallbox",
			Type:     "cloud",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &WallboxChargerPoller{
				Device: device,
				httpClient: &http.Client{
					Timeout: 10 * time.Second,
				},
			}
		},
	})
}

func (p *WallboxChargerPoller) Connect() error {
	if p.Device.Username == "" || p.Device.Password == "" {
		p.status = "error"
		return fmt.Errorf("wallbox credentials missing")
	}

	req, _ := http.NewRequest("GET", "https://api.wall-box.com/auth/token/user", nil)
	req.SetBasicAuth(p.Device.Username, p.Device.Password)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.status = "error"
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.status = "error"
		return fmt.Errorf("wallbox auth failed: %s", resp.Status)
	}

	var authResp struct {
		Jwt string `json:"jwt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		p.status = "error"
		return err
	}
	p.token = authResp.Jwt

	// Get chargers
	req2, _ := http.NewRequest("GET", "https://api.wall-box.com/v3/chargers/groups", nil)
	req2.Header.Set("Authorization", "Bearer "+p.token)
	resp2, err := p.httpClient.Do(req2)
	if err == nil && resp2.StatusCode == http.StatusOK {
		defer resp2.Body.Close()
		var groupsResp struct {
			Result []struct {
				Chargers []struct {
					Id int `json:"id"`
				} `json:"chargers"`
			} `json:"result"`
		}
		json.NewDecoder(resp2.Body).Decode(&groupsResp)
		if len(groupsResp.Result) > 0 && len(groupsResp.Result[0].Chargers) > 0 {
			p.chargerID = groupsResp.Result[0].Chargers[0].Id
			p.status = "online"
			return nil
		}
	}

	p.status = "error"
	return fmt.Errorf("wallbox no chargers found")
}

func (p *WallboxChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *WallboxChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.token == "" || p.chargerID == 0 {
		return 0, 0, 0, 0, 0, nil
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.wall-box.com/chargers/status/%d", p.chargerID), nil)
	req.Header.Set("Authorization", "Bearer "+p.token)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		p.Connect() // re-auth
		return 0, 0, 0, 0, 0, nil
	}

	var stateResp struct {
		ChargingPower float64 `json:"charging_power"`
		AddedEnergy   float64 `json:"added_energy"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &stateResp)

	// Wallbox returns power in W and energy in Wh
	return stateResp.ChargingPower, 0, 0, stateResp.AddedEnergy / 1000.0, 0, nil
}

func (p *WallboxChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *WallboxChargerPoller) SetChargeCurrent(amps float64) error {
	if p.status != "online" || p.token == "" || p.chargerID == 0 {
		return nil
	}

	data := map[string]interface{}{
		"maxChargingCurrent": amps,
	}
	bodyBytes, _ := json.Marshal(data)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("https://api.wall-box.com/v2/charger/%d", p.chargerID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	return err
}

func (p *WallboxChargerPoller) Close() error {
	return nil
}
