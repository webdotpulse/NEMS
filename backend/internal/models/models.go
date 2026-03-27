package models

type SiteSettings struct {
	StrategyMode              string  `json:"strategy_mode"`
	CapacityPeakLimitKw       float64 `json:"capacity_peak_limit_kw"`
	ActiveInverterCurtailment bool    `json:"active_inverter_curtailment"`
	ForceChargeBelowEuro      float64 `json:"force_charge_below_euro"`
	SmartEvCheapestHours      int     `json:"smart_ev_cheapest_hours"`
	GridNominalCurrentA       float64 `json:"grid_nominal_current_a"`
	GridSystem                string  `json:"grid_system"`
	AllowedGridImportKw       float64 `json:"allowed_grid_import_kw"`
	AllowedGridExportKw       float64 `json:"allowed_grid_export_kw"`
}

type Device struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Template        string  `json:"template"`
	Host            string  `json:"host"`
	Port            int     `json:"port"`
	ModbusID        int     `json:"modbus_id"`
	Username        string  `json:"username,omitempty"`
	Password        string  `json:"password,omitempty"`
	Status          string  `json:"status"`
	HasGridMeter    bool    `json:"has_grid_meter"`
	HasBattery      bool    `json:"has_battery"`
	BatteryCapacity float64 `json:"battery_capacity"`
}

type DevicePoller interface {
	Connect() error
	Poll() (powerW float64, batteryPowerW float64, gridPowerW float64, energyKwh float64, soc float64, err error)
	GetDevice() Device
	Status() string
	Close() error
}

type ChargeController interface {
	SetChargeCurrent(amps float64) error
}

type BatteryController interface {
	DischargeBattery(powerW float64) error
	ChargeBattery(powerW float64) error
}

type InverterController interface {
	SetActivePowerLimit(powerW float64) error
}
