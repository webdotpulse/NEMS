# Pulse EMS User Manual

Welcome to the comprehensive user manual for Pulse EMS (Energy Management System). Pulse EMS is a lightweight, highly responsive, and fully UI-driven Energy Management System optimized for low-power devices like the Raspberry Pi. This manual details how to set up, configure, and optimize your smart home energy usage using the available features and parameters.

---

## 1. Introduction & Architecture Overview

Pulse EMS serves as the brain of your home energy ecosystem, integrating Grid, Solar, Battery, and EV Charger hardware. By seamlessly monitoring realtime power usage and acting upon defined optimization strategies, it helps reduce electricity costs, maximize self-consumption, and protect you from high capacity grid tariffs.

### Core Philosophy
* **Fully UI-Driven:** No YAML or configuration file editing is required. Every hardware device and optimization rule is configured straight from the frontend interface.
* **Responsive Architecture:** A highly optimized Go backend paired with an embedded SQLite database and a modern Vue 3 SPA guarantees snappy updates using Server-Sent Events (SSE).
* **Minimal Wear:** Database writes are batched and buffered to maximize the lifespan of your device's SD Card.
* **Privacy First:** Your data never leaves your home unless you use external proxy or API configurations.

---

## 2. Installation & Setup

Pulse EMS is primarily deployed via release artifacts.

### Recommended: Custom OS Image
The easiest way to install Pulse EMS is by flashing the pre-built **Custom Raspberry Pi OS Lite (ARM64)** image to an SD card.
1. Download the `nems-os-image.img.xz` from the GitHub releases page.
2. Flash it using an imaging tool like BalenaEtcher or Raspberry Pi Imager.
3. Insert into the Raspberry Pi and boot.
4. Open your web browser to `http://ems` or `http://ems.local` (or your device IP on port `8080`).

### Alternative: Debian Package
If running an existing Debian or Ubuntu OS on an ARM64 system:
```bash
sudo dpkg -i nems_*_arm64.deb
sudo apt-get install -f
```
The `nems.service` will start automatically.

---

## 3. Dashboard & Real-Time Monitoring

The primary **Dashboard** visualizes your energy distribution in real-time.

* **Power Flow Interactive Graphic:** Displays active power nodes (Grid, Solar, Battery, Charger, Home). Nodes are intelligently hidden if not configured.
* **Click-to-Reveal:** Clicking any individual node in the power flow diagram reveals detailed historical line charts and statistics unique to that device category.
* **Metrics Board:** Provides an aggregated daily view of solar yield, grid import/export totals, and self-consumption ratios.

---

## 4. Global Settings & Configurations

Access the **Settings** view from the top navigation bar to configure system-wide rules and energy contracts. The parameters map directly to your system's underlying capabilities.

### 4.1 Site Optimization (Strategy Mode)
Located under Site Optimization, these parameters dictate how the EMS prioritizes energy distribution.

* **`strategy_mode`**: Selects the primary operational behavior of the system.
  * **Eco:** Prioritizes charging the battery and EV with excess solar. Limits reliance on grid import.
  * **Flanders:** Activates predictive Peak Shaving based on the Belgian/Flanders capacity tariff model. Calculates rolling 15-minute average peaks and dynamically throttles appliances.
  * **Netherlands:** Focuses on minimizing or completely eliminating solar feed-in to the grid (zero-export).

* **`capacity_peak_limit_kw`**: The absolute maximum average quarter-hour grid import allowed (typically used in Flanders mode).
* **`peak_shaving_buffer_w`**: The safety buffer applied when nearing the capacity peak limit to prevent overshoot.
* **`peak_shaving_rampup_w`**: Defines how fast a throttled device (like an EV charger) is allowed to increase power when capacity frees up.
* **`grid_nominal_current_a`**: Your household's main grid connection amperage (e.g., `25`, `32`, `40`). Protects the main fuse by throttling chargers if the sum load exceeds this limit.
* **`grid_system`**: E.g., `single_phase_230v` or `three_phase_400v`. Adjusts power-to-amps calculations.
* **`allowed_grid_import_kw`**: Hard limit on grid import. Devices will throttle to respect this.
* **`allowed_grid_export_kw`**: Hard limit on solar feed-in. If this is zero, zero-export logic applies.
* **`active_inverter_curtailment`**: Allows the system to actively throttle solar inverters to respect the allowed grid export limits.
* **`appliance_turn_on_excess_w`**: Amount of continuous solar export required before triggering smart appliances or relays.

### 4.2 Battery Arbitrage & Schedules
Take advantage of variable electricity pricing to charge from the grid when cheap and use the battery when expensive.

* **`battery_grid_charge_strategy`**:
  * `price_only`: Charges based on simple EPEX spot price thresholds.
  * `super_dal_only`: Restricts grid charging to specific highly optimized contract windows (like Engie Superdal).
  * `hybrid`: Combines price-based charging with Superdal optimization.
  * `dynamic_forecast`: Maps optimal hourly charge/discharge behavior based on solar weather forecasts and the home's baseline load.

* **`force_charge_below_euro`**: Force charges the battery from the grid when the spot price drops below this value (€/kWh).
* **`force_discharge_above_euro`**: Force discharges the battery to the grid when the spot price spikes above this value.
* **`superdal_optimization_enabled`**: Enables specific provider logic, e.g. charging exactly during specific hours based on cheap tariff slots.
* **`superdal_target_soc`**: The desired State of Charge to reach by the end of the optimized time window.
* **`custom_charge_schedule`**: Configurable forced-charge windows defined by a start time, end time, and target State of Charge (SOC).

### 4.3 Energy Contracts & Pricing
Visualizes and modifies how your electricity price is calculated per kWh based on Day-Ahead (EPEX) spot prices or fixed rates.

* **`contract_type`**: `dynamic` (market spot prices) or `fixed` (static peak/off-peak).
* **`fixed_price_peak_kwh` / `fixed_price_off_peak_kwh`**: Rates applied if `contract_type` is `fixed`.
* **`fixed_inject_price_kwh`**: Rate you receive for feeding into the grid on a fixed contract.
* **Dynamic Contract Variables**:
  * **`dynamic_markup_kwh`**: A simple flat markup applied to the base EPEX spot price.
  * **`dynamic_inject_multiplier`**: Multiplier applied to spot prices for energy injection (usually to calculate provider fees on return).
  * **Provider Configurations (Engie, Luminus, Eneco, Frank, Ecopower)**: Includes unique combinations of base fees (e.g. `engie_base_fee`), markup structures, and pricing multipliers (`*_multiplier`, `*_inject_multiplier`) allowing exact mapping of your energy bill parameters to the UI calculations.

### 4.4 Smart Modes
Advanced configuration for optimizing your home consumption.

* **`smart_ev_cheapest_hours`**: Automatically identifies the $N$ cheapest hours of the day to charge your Electric Vehicle.
* **`home_base_load_w`**: Estimated average background consumption of the house used for predicting future battery behavior in the `dynamic_forecast` mode.

---

## 5. Device Management & Configuration

To integrate hardware, use the **Add Device** feature.

* **Templates**: NEMS supports hardware from Huawei, SMA, Solis, Raedian, Easee, Enerlution, and more. Select the template that matches your hardware type.
* **Network Scanner**: A zero-dependency network discovery tool is available in the UI. It scans your local subnet and matches MAC Organizationally Unique Identifiers (OUIs) to known hardware, assisting you in finding device IPs instantly.
* **Host & Port**: The IP address of the device on your local network (e.g., `192.168.1.50`) and communication port (usually `502` for Modbus).
* **`modbus_id`**: The Modbus Slave ID of the device (often `1` or `2`).
* **`poll_interval`**: How often (in seconds) the system queries the device. The default is `5` seconds. High-priority devices like grid meters can be set lower for faster reaction times.
* **`charge_mode` (EV Chargers)**: Defines how the charger should act: e.g., `eco` (solar only) vs `fast` (grid + solar).
* **`battery_mode` (Batteries)**: `auto` (follows strategy), `force_charge`, or `force_discharge`.

---

## 6. Native OCPP Server

Pulse EMS features a built-in Native OCPP (Open Charge Point Protocol) server for seamless communication with standard EV Chargers (1.6 / 2.0.1).

* **Configuration**: In your EV Charger's dedicated app or web interface, configure the CSMS / Backend URL to point to the EMS IP address.
* **URL Format**: The endpoint usually resembles `ws://<EMS-IP>:8080/api/ocpp/<Charger-ID>`. Check the dedicated OCPP setup interface in the EMS for your exact customized URL.
* Once connected, the charger acts as any other local device, reporting power, receiving setpoints, and integrating into the EMS ecosystem without external cloud dependency.

---

## 7. System Updates

Pulse EMS supports simple, single-click over-the-air updates.

* **Update Check:** Navigating to the System Info panel will automatically query GitHub for new releases.
* **Token Configuration (`github_token`)**: A GitHub token can be configured in settings to avoid IP rate limiting from GitHub's API during update checks.
* **Process**: Clicking "Install Update" downloads the latest `.deb` package to the system and executes it transparently. The UI will show realtime installation logs. Once completed, the backend service restarts automatically, and the UI will reconnect.
