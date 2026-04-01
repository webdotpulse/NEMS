package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"nems/internal/models"
)

type ZaptecChargerPoller struct {
	Device     models.Device
	status     string
	token      string
	chargerID  string
	httpClient *http.Client
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "zaptec_charger",
			Name:     "Zaptec Pro/Go",
			Vendor:   "Zaptec",
			Type:     "cloud",
			Category: "evse",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &ZaptecChargerPoller{
				Device: device,
				httpClient: &http.Client{
					Timeout: 10 * time.Second,
				},
			}
		},
	})
}

func (p *ZaptecChargerPoller) Connect() error {
	if p.Device.Username == "" || p.Device.Password == "" {
		p.status = "error"
		return fmt.Errorf("zaptec credentials missing")
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", p.Device.Username)
	data.Set("password", p.Device.Password)

	req, err := http.NewRequest("POST", "https://api.zaptec.com/oauth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		p.status = "error"
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.status = "error"
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.status = "error"
		return fmt.Errorf("zaptec auth failed: %s", resp.Status)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		p.status = "error"
		return err
	}
	p.token = authResp.AccessToken

	// Fetch chargers to get ID
	req2, _ := http.NewRequest("GET", "https://api.zaptec.com/api/chargers", nil)
	req2.Header.Set("Authorization", "Bearer "+p.token)
	resp2, err := p.httpClient.Do(req2)
	if err == nil && resp2.StatusCode == http.StatusOK {
		defer resp2.Body.Close()
		var chargersResp struct {
			Data []struct {
				Id string `json:"Id"`
			} `json:"Data"`
		}
		json.NewDecoder(resp2.Body).Decode(&chargersResp)
		if len(chargersResp.Data) > 0 {
			p.chargerID = chargersResp.Data[0].Id
			p.status = "online"
			return nil
		}
	}

	p.status = "error"
	return fmt.Errorf("zaptec no chargers found")
}

func (p *ZaptecChargerPoller) Status() string {
	if p.status == "" {
		return "offline"
	}
	return p.status
}

func (p *ZaptecChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	if p.status != "online" || p.token == "" || p.chargerID == "" {
		return 0, 0, 0, 0, 0, nil
	}

	req, _ := http.NewRequest("GET", "https://api.zaptec.com/api/chargers/"+p.chargerID+"/state", nil)
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

	var stateResp []struct {
		StateId int    `json:"StateId"`
		Value   string `json:"ValueAsString"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &stateResp)

	powerW := 0.0
	energyKwh := 0.0

	for _, obs := range stateResp {
		if obs.StateId == 703 { // TotalChargePower
			fmt.Sscanf(obs.Value, "%f", &powerW)
		} else if obs.StateId == 710 { // TotalChargePowerSession
			fmt.Sscanf(obs.Value, "%f", &energyKwh)
		}
	}

	// Zaptec returns power in kW
	return powerW * 1000.0, 0, 0, energyKwh, 0, nil
}

func (p *ZaptecChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *ZaptecChargerPoller) SetChargeCurrent(amps float64) error {
	if p.status != "online" || p.token == "" || p.chargerID == "" {
		return nil
	}

	data := map[string]interface{}{
		"MaxChargeCurrent": amps,
	}
	bodyBytes, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", "https://api.zaptec.com/api/chargers/"+p.chargerID+"/update", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}
	return err
}

func (p *ZaptecChargerPoller) Close() error {
	return nil
}
