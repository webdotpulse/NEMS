export interface Device {
  id: number;
  name: string;
  template: string;
  host: string;
  port: number;
  modbus_id: number;
  username?: string;
  password?: string;
  charge_mode?: string;
  status?: string;
  has_grid_meter?: boolean;
  has_battery?: boolean;
  battery_capacity?: number;
}

export interface SiteSettings {
  strategy_mode: string;
  capacity_peak_limit_kw: number;
  active_inverter_curtailment: boolean;
  force_charge_below_euro: number;
  smart_ev_cheapest_hours: number;
  grid_nominal_current_a: number;
  grid_system: string;
  allowed_grid_import_kw: number;
  allowed_grid_export_kw: number;
  appliance_turn_on_excess_w: number;
  peak_shaving_buffer_w: number;
  peak_shaving_rampup_w: number;
}

export interface SiteState {
  grid_power_w: number | null;
  solar_power_w: number | null;
  battery_power_w: number | null;
  battery_soc: number | null;
  total_load_w: number | null;
  ev_charger_power_w: number | null;
  device_health?: Record<number, string>;
}

export interface DailyAggregates {
  grid_import_kwh: number;
  grid_export_kwh: number;
  solar_yield_kwh: number;
  battery_charge_kwh: number;
  battery_charge_solar_kwh: number;
  battery_charge_grid_kwh: number;
  battery_discharge_kwh: number;
  house_consumption_kwh: number;
}

export interface Template {
  id: string;
  name: string;
  type: string;
}

export interface SystemInfo {
  hostname: string;
  ip: string;
  netmask: string;
}
