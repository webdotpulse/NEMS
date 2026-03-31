package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"nems/internal/models"
	"nems/internal/templates"
)

// handleLogs is the HTTP handler for the /api/logs endpoint.
func handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "DELETE" {
		logBuffer.ClearLogs()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "cleared"}`))
		return
	}
	logs := logBuffer.GetLogs()
	json.NewEncoder(w).Encode(logs)
}

// handleSystemInfo is the HTTP handler for the /api/system/info endpoint.
func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	var primaryIP string
	var primaryNetmask string

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					primaryIP = ipnet.IP.String()
					mask := ipnet.Mask
					primaryNetmask = net.IP(mask).String()
					break
				}
			}
		}
	}

	// Default values
	gateway := "unknown"
	memInfo := "unknown"
	diskInfo := "unknown"

	// Get gateway
	out, err := exec.Command("ip", "route").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "default via") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					gateway = parts[2]
					break
				}
			}
		}
	}

	// Get Memory Info
	memFile, err := os.Open("/proc/meminfo")
	if err == nil {
		defer memFile.Close()
		scanner := bufio.NewScanner(memFile)
		var totalMem, freeMem int64
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d kB", &totalMem)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d kB", &freeMem)
			}
		}
		if totalMem > 0 {
			usedMem := totalMem - freeMem
			memInfo = fmt.Sprintf("%.1f GB / %.1f GB", float64(usedMem)/(1024*1024), float64(totalMem)/(1024*1024))
		}
	}

	// Get Disk Info
	diskOut, err := exec.Command("df", "-h", "/").Output()
	if err == nil {
		lines := strings.Split(string(diskOut), "\n")
		if len(lines) >= 2 {
			parts := strings.Fields(lines[1])
			if len(parts) >= 5 {
				diskInfo = fmt.Sprintf("%s / %s (%s used)", parts[2], parts[1], parts[4])
			}
		}
	}

	info := map[string]string{
		"hostname": hostname,
		"ip":       primaryIP,
		"netmask":  primaryNetmask,
		"gateway":  gateway,
		"memory":   memInfo,
		"disk":     diskInfo,
		"build":    BuildNumber,
	}

	json.NewEncoder(w).Encode(info)
}

// handleSystemResetDb is the HTTP handler for the /api/system/reset-db endpoint.
func handleSystemResetDb(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err := db.Exec("DELETE FROM measurements")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM devices")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE site_settings SET strategy_mode='eco', capacity_peak_limit_kw=2.5, active_inverter_curtailment=0, force_charge_below_euro=0.0, force_discharge_above_euro=999.0, smart_ev_cheapest_hours=0, grid_nominal_current_a=25.0, grid_system='single_phase_230v', allowed_grid_import_kw=0.0, allowed_grid_export_kw=0.0, appliance_turn_on_excess_w=0.0, peak_shaving_buffer_w=200.0, peak_shaving_rampup_w=500.0")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if PollerMgr != nil {
		PollerMgr.SyncDevices()
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "database reset"}`))
}

// handleSystemReboot is the HTTP handler for the /api/system/reboot endpoint.
func handleSystemReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "rebooting"}`))

	// Give the HTTP response time to be sent before rebooting
	go func() {
		time.Sleep(2 * time.Second)
		out, err := exec.Command("sudo", "reboot").CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Reboot failed: %v, output: %s", err, string(out))
		}
	}()
}

// handleStatus is the HTTP handler for the /api/status endpoint.
// It returns a basic JSON object indicating the API is running.
func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

// handleTariffsToday is the HTTP handler for the /api/tariffs/today endpoint.
// It retrieves EPEX spot prices from the local database for today and tomorrow.
func handleTariffsToday(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
	endOfTomorrow := time.Date(now.Year(), now.Month(), now.Day()+2, 0, 0, 0, 0, loc).UTC()

	rows, err := db.Query("SELECT timestamp, price_per_kwh FROM epex_prices WHERE timestamp >= ? AND timestamp < ? ORDER BY timestamp ASC", startOfToday, endOfTomorrow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var prices []PricePoint
	for rows.Next() {
		var p PricePoint
		var ts time.Time
		if err := rows.Scan(&ts, &p.PricePerKwh); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.Timestamp = ts.In(loc)
		prices = append(prices, p)
	}

	if prices == nil {
		prices = []PricePoint{}
	}
	json.NewEncoder(w).Encode(prices)
}

// handleSettings is the HTTP handler for the /api/settings endpoint.
// It supports GET requests to retrieve site optimization settings,
// and PUT requests to update them.
func handleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, force_charge_below_euro, force_discharge_above_euro, smart_ev_cheapest_hours, grid_nominal_current_a, grid_system, allowed_grid_import_kw, allowed_grid_export_kw, appliance_turn_on_excess_w, peak_shaving_buffer_w, peak_shaving_rampup_w, timezone, latitude, longitude FROM site_settings WHERE id = 1")
		var settings models.SiteSettings
		err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.ForceChargeBelowEuro, &settings.ForceDischargeAboveEuro, &settings.SmartEvCheapestHours, &settings.GridNominalCurrentA, &settings.GridSystem, &settings.AllowedGridImportKw, &settings.AllowedGridExportKw, &settings.ApplianceTurnOnExcessW, &settings.PeakShavingBufferW, &settings.PeakShavingRampupW, &settings.Timezone, &settings.Latitude, &settings.Longitude)
		if err != nil {
			if err == sql.ErrNoRows {
				// Fallback
				settings = models.SiteSettings{
					StrategyMode:              "eco",
					CapacityPeakLimitKw:       2.5,
					ActiveInverterCurtailment: false,
					ForceChargeBelowEuro:      0.0,
					ForceDischargeAboveEuro:   999.0,
					SmartEvCheapestHours:      0,
					GridNominalCurrentA:       25.0,
					GridSystem:                "single_phase_230v",
					AllowedGridImportKw:       0.0,
					AllowedGridExportKw:       0.0,
					ApplianceTurnOnExcessW:    0.0,
					PeakShavingBufferW:        200.0,
					PeakShavingRampupW:        500.0,
					Timezone:                  "Europe/Brussels",
					Latitude:                  50.8503,
					Longitude:                 4.3517,
				}
				json.NewEncoder(w).Encode(settings)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(settings)
	} else if r.Method == "PUT" {
		var settings models.SiteSettings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if settings.Timezone == "" {
			settings.Timezone = "Europe/Brussels"
		}

		_, err := db.Exec("UPDATE site_settings SET strategy_mode = ?, capacity_peak_limit_kw = ?, active_inverter_curtailment = ?, force_charge_below_euro = ?, force_discharge_above_euro = ?, smart_ev_cheapest_hours = ?, grid_nominal_current_a = ?, grid_system = ?, allowed_grid_import_kw = ?, allowed_grid_export_kw = ?, appliance_turn_on_excess_w = ?, peak_shaving_buffer_w = ?, peak_shaving_rampup_w = ?, timezone = ?, latitude = ?, longitude = ? WHERE id = 1",
			settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment, settings.ForceChargeBelowEuro, settings.ForceDischargeAboveEuro, settings.SmartEvCheapestHours, settings.GridNominalCurrentA, settings.GridSystem, settings.AllowedGridImportKw, settings.AllowedGridExportKw, settings.ApplianceTurnOnExcessW, settings.PeakShavingBufferW, settings.PeakShavingRampupW, settings.Timezone, settings.Latitude, settings.Longitude)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(settings)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTemplates is the HTTP handler for the /api/templates endpoint.
// It returns a list of available device templates for frontend configuration.
func handleTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tList := templates.GetTemplates()
	json.NewEncoder(w).Encode(tList)
}

// handleDevices is the HTTP handler for the /api/devices endpoint.
// It supports GET requests to retrieve all configured hardware devices,
// and POST requests to add a new device.
func handleDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, charge_mode, battery_mode, ocpp_proxy_url FROM devices")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var devices []models.Device
		for rows.Next() {
			var d models.Device
			var username sql.NullString
			var password sql.NullString
			var chargeMode sql.NullString
			var batteryMode sql.NullString
			var proxyUrl sql.NullString
			if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password, &d.HasGridMeter, &d.HasBattery, &d.BatteryCapacity, &chargeMode, &batteryMode, &proxyUrl); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if username.Valid {
				d.Username = username.String
			}
			if password.Valid {
				d.Password = password.String
			}
			if chargeMode.Valid {
				d.ChargeMode = chargeMode.String
			}
			if batteryMode.Valid {
				d.BatteryMode = batteryMode.String
			}
			if proxyUrl.Valid {
				d.OcppProxyUrl = proxyUrl.String
			}

			// Set dynamic status if poller exists
			d.Status = "offline"
			if PollerMgr != nil {
				PollerMgr.mu.Lock()
				if poller, ok := PollerMgr.pollers[d.ID]; ok {
					d.Status = poller.Status()
				}
				PollerMgr.mu.Unlock()
			}

			d.Category = templates.GetCategory(d.Template)

			devices = append(devices, d)
		}
		// ensure non-nil slice in json
		if devices == nil {
			devices = []models.Device{}
		}
		json.NewEncoder(w).Encode(devices)
	} else if r.Method == "POST" {
		var d models.Device
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if d.ChargeMode == "" {
			d.ChargeMode = "eco"
		}
		if d.BatteryMode == "" {
			d.BatteryMode = "auto"
		}

		result, err := db.Exec("INSERT INTO devices (name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, charge_mode, battery_mode, ocpp_proxy_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.ChargeMode, d.BatteryMode, d.OcppProxyUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		d.ID = int(id)

		// Notify poller manager
		if PollerMgr != nil {
			PollerMgr.SyncDevices()
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(d)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDevice is the HTTP handler for the /api/devices/{id} and /api/devices/{id}/mode endpoints.
// It supports PUT requests to edit a device or update its mode,
// and DELETE requests to remove a device.
func handleDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/api/devices/")
	if strings.HasSuffix(idStr, "/mode") {
		idStr = strings.TrimSuffix(idStr, "/mode")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	if r.Method == "DELETE" {
		_, err := db.Exec("DELETE FROM devices WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Notify poller manager
		if PollerMgr != nil {
			PollerMgr.SyncDevices()
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "deleted"}`))
	} else if strings.HasSuffix(r.URL.Path, "/mode") && r.Method == "PUT" {
		var payload struct {
			ChargeMode  string `json:"charge_mode"`
			BatteryMode string `json:"battery_mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if payload.ChargeMode != "" {
			_, err = db.Exec("UPDATE devices SET charge_mode = ? WHERE id = ?", payload.ChargeMode, id)
		} else if payload.BatteryMode != "" {
			_, err = db.Exec("UPDATE devices SET battery_mode = ? WHERE id = ?", payload.BatteryMode, id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "updated"}`))

	} else if r.Method == "PUT" {
		var d models.Device
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if d.ChargeMode == "" {
			d.ChargeMode = "eco"
		}
		if d.BatteryMode == "" {
			d.BatteryMode = "auto"
		}

		_, err = db.Exec("UPDATE devices SET name = ?, template = ?, host = ?, port = ?, modbus_id = ?, username = ?, password = ?, has_grid_meter = ?, has_battery = ?, battery_capacity = ?, charge_mode = ?, battery_mode = ?, ocpp_proxy_url = ? WHERE id = ?",
			d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.ChargeMode, d.BatteryMode, d.OcppProxyUrl, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		d.ID = id

		// Notify poller manager
		if PollerMgr != nil {
			PollerMgr.SyncDevices()
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(d)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
