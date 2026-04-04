package models

type SiteSettings struct {
	StrategyMode              string  `json:"strategy_mode"`
	CapacityPeakLimitKw       float64 `json:"capacity_peak_limit_kw"`
	ActiveInverterCurtailment bool    `json:"active_inverter_curtailment"`
	BatteryGridChargeStrategy string  `json:"battery_grid_charge_strategy"`
	ForceChargeBelowEuro      float64 `json:"force_charge_below_euro"`
	ForceDischargeAboveEuro   float64 `json:"force_discharge_above_euro"`
	SmartEvCheapestHours      int     `json:"smart_ev_cheapest_hours"`
	GridNominalCurrentA       float64 `json:"grid_nominal_current_a"`
	GridSystem                string  `json:"grid_system"`
	AllowedGridImportKw       float64 `json:"allowed_grid_import_kw"`
	AllowedGridExportKw       float64 `json:"allowed_grid_export_kw"`
	ApplianceTurnOnExcessW    float64 `json:"appliance_turn_on_excess_w"`
	PeakShavingBufferW        float64 `json:"peak_shaving_buffer_w"`
	PeakShavingRampupW        float64 `json:"peak_shaving_rampup_w"`
	Timezone                  string  `json:"timezone"`
	Language                  string  `json:"language"`
	Address                   string  `json:"address"`
	Latitude                  float64 `json:"latitude"`
	Longitude                 float64 `json:"longitude"`

	// Energy Contract Configuration
	ContractType                 string  `json:"contract_type"`
	ScaleFactorEpexSpotConsumption float64 `json:"scale_factor_epex_spot_consumption"`
	EnergyPricesConsumption        float64 `json:"energy_prices_consumption"`
	GridCostsConsumption           float64 `json:"grid_costs_consumption"`
	ScaleFactorEpexSpotInjection   float64 `json:"scale_factor_epex_spot_injection"`
	EnergyPricesInjection          float64 `json:"energy_prices_injection"`
	GridCostsInjection             float64 `json:"grid_costs_injection"`
	EngieMarkupPeak              float64 `json:"engie_markup_peak"`
	EngieMarkupOffPeak      float64 `json:"engie_markup_off_peak"`
	EngieMarkupSuperOffPeak float64 `json:"engie_markup_super_off_peak"`
	EngieMultiplier         float64 `json:"engie_multiplier"`
	EngieBaseFee            float64 `json:"engie_base_fee"`

	// Custom Schedules & Optimization
	CustomChargeSchedule        string  `json:"custom_charge_schedule"`
	SuperdalOptimizationEnabled bool    `json:"superdal_optimization_enabled"`
	SuperdalTargetSoc           float64 `json:"superdal_target_soc"`

	// System Update
	GithubToken string `json:"github_token"`
}

type Device struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Template        string  `json:"template"`
	Category        string  `json:"category,omitempty"`
	Host            string  `json:"host"`
	Port            int     `json:"port"`
	ModbusID        int     `json:"modbus_id"`
	Username        string  `json:"username,omitempty"`
	Password        string  `json:"password,omitempty"`
	ChargeMode      string  `json:"charge_mode,omitempty"`
	BatteryMode     string  `json:"battery_mode,omitempty"`
	Status          string  `json:"status"`
	HasGridMeter         bool    `json:"has_grid_meter"`
	HasBattery           bool    `json:"has_battery"`
	BatteryCapacity      float64 `json:"battery_capacity"`
	InverterRatedPowerKw float64 `json:"inverter_rated_power_kw,omitempty"`
	PollInterval         int     `json:"poll_interval"`
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

type RelayController interface {
	SetState(on bool) error
}
