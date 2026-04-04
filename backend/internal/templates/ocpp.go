package templates

import (
	"log"
	"time"

	"nems/internal/models"
	"nems/internal/ocpp"
)

type OcppChargerPoller struct {
	Device models.Device
	status string
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:       "ocpp_charger",
			Name:     "OCPP Charger",
			Vendor:   "Generic",
			Type:     "ocpp", // Allow custom ocpp fields configuration
			Category: "charger",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &OcppChargerPoller{Device: device}
		},
	})
}

func (p *OcppChargerPoller) Connect() error {
	// The charger connects to us. We just use the host field as the ChargePointId.
	log.Printf("[INFO] OcppChargerPoller: Initialized for ChargePointId %s", p.Device.Host)
	// The EV charger must be configured to point its CSMS URL to ws://<nems-ip>:<port>/api/ocpp/<ChargePointId>
	p.status = "online"
	return nil
}

func (p *OcppChargerPoller) Status() string {
	// Check if the WebSocket server has an active connection for this charge point
	state := ocpp.GetChargerState(p.Device.Host)
	if state != nil && state.IsConnected() {
		_, _, lastSeen := state.GetTelemetry()
		if time.Since(lastSeen) < 5*time.Minute {
			return "online"
		}
	}
	return "offline"
}

func (p *OcppChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	state := ocpp.GetChargerState(p.Device.Host)
	if state == nil {
		return 0, 0, 0, 0, 0, nil
	}

	powerW, energyWh, _ := state.GetTelemetry()
	energyKwh := energyWh / 1000.0

	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *OcppChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *OcppChargerPoller) SetChargeCurrent(amps float64) error {
	state := ocpp.GetChargerState(p.Device.Host)
	if state != nil {
		return state.SetChargingProfile(amps)
	}
	log.Printf("[WARN] OcppChargerPoller: Cannot set current, charger %s offline", p.Device.Host)
	return nil
}

func (p *OcppChargerPoller) Close() error {
	log.Printf("[INFO] OcppChargerPoller: Closing connection")
	p.status = "offline"
	return nil
}
