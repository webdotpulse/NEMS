package templates

import (
	"math/rand"

	"nems/internal/models"
)

type DemoInverterPoller struct {
	Device models.Device
}

func init() {
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "demo_inverter",
			Name: "Demo Inverter",
			Type: "demo",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &DemoInverterPoller{Device: device}
		},
	})
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "demo_dongle",
			Name: "Demo Grid Meter",
			Type: "demo",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &DemoDonglePoller{Device: device}
		},
	})
	Register(Template{
		Metadata: TemplateMetadata{
			ID:   "demo_charger",
			Name: "Demo EV Charger",
			Type: "demo",
		},
		NewPoller: func(device models.Device) models.DevicePoller {
			return &DemoChargerPoller{Device: device}
		},
	})
}

func (p *DemoInverterPoller) Connect() error {
	return nil
}

func (p *DemoInverterPoller) Status() string {
	return "online"
}

func (p *DemoInverterPoller) Poll() (float64, float64, float64, float64, float64, error) {
	powerW := 1000.0 + rand.Float64()*3000.0
	batteryPowerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, batteryPowerW, 0, energyKwh, 0, nil
}

func (p *DemoInverterPoller) GetDevice() models.Device {
	return p.Device
}

func (p *DemoInverterPoller) Close() error {
	return nil
}

type DemoDonglePoller struct {
	Device models.Device
}

func (p *DemoDonglePoller) Connect() error {
	return nil
}

func (p *DemoDonglePoller) Status() string {
	return "online"
}

func (p *DemoDonglePoller) Poll() (float64, float64, float64, float64, float64, error) {
	powerW := -2000.0 + rand.Float64()*4000.0
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return 0, 0, powerW, energyKwh, 0, nil
}

func (p *DemoDonglePoller) GetDevice() models.Device {
	return p.Device
}

func (p *DemoDonglePoller) Close() error {
	return nil
}

type DemoChargerPoller struct {
	Device models.Device
}

func (p *DemoChargerPoller) Connect() error {
	return nil
}

func (p *DemoChargerPoller) Status() string {
	return "online"
}

func (p *DemoChargerPoller) Poll() (float64, float64, float64, float64, float64, error) {
	powerW := 0.0
	if rand.Float32() > 0.5 {
		powerW = 11000.0
	}
	energyKwh := powerW * (5.0 / 3600.0) / 1000.0
	return powerW, 0, 0, energyKwh, 0, nil
}

func (p *DemoChargerPoller) GetDevice() models.Device {
	return p.Device
}

func (p *DemoChargerPoller) Close() error {
	return nil
}
