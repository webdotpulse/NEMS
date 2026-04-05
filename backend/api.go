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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	assetTag := "unknown"

	// Get Asset Tag (Serial from /proc/cpuinfo)
	cpuFile, err := os.Open("/proc/cpuinfo")
	if err == nil {
		defer cpuFile.Close()
		scanner := bufio.NewScanner(cpuFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Serial") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					assetTag = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

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
		"hostname":  hostname,
		"ip":        primaryIP,
		"netmask":   primaryNetmask,
		"gateway":   gateway,
		"memory":    memInfo,
		"disk":      diskInfo,
		"build":     BuildNumber,
		"asset_tag": assetTag,
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

var (
	updateState string // "idle", "installing", "done", "error"
	updateLogs  []string
	updateMu    sync.Mutex
)

func logUpdate(msg string, a ...interface{}) {
	text := fmt.Sprintf(msg, a...)
	log.Printf("[UPDATE] %s", text)

	updateMu.Lock()
	defer updateMu.Unlock()
	updateLogs = append(updateLogs, text)
	// Keep log size bounded
	if len(updateLogs) > 100 {
		updateLogs = updateLogs[len(updateLogs)-100:]
	}
}

// handleSystemUpdateStatus returns the current status and logs of the system update.
func handleSystemUpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateMu.Lock()
	defer updateMu.Unlock()

	response := map[string]interface{}{
		"status": updateState,
		"logs":   updateLogs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSystemUpdateCheck checks the GitHub releases API for a newer version.
func handleSystemUpdateCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Call GitHub API to get the latest releases
	req, err := http.NewRequest("GET", "https://api.github.com/repos/webdotpulse/NEMS/releases", nil)
	if err != nil {
		http.Error(w, `{"error": "Failed to create request"}`, http.StatusInternalServerError)
		return
	}

	var githubToken string
	db.QueryRow("SELECT github_token FROM site_settings WHERE id = 1").Scan(&githubToken)
	if githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+githubToken)
	}

	// Add a User-Agent header, GitHub API requires it
	req.Header.Set("User-Agent", "NEMS-Update-Checker")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch latest release"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Log the status and body for debugging
		log.Printf("[ERROR] GitHub API returned status: %d", resp.StatusCode)
		// Instead of returning 500, we could just return update available = false
		// since we cannot get the version.
		response := map[string]interface{}{
			"update_available": false,
			"latest_version":   "unknown",
			"current_version":  BuildNumber,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		http.Error(w, `{"error": "Failed to parse release info"}`, http.StatusInternalServerError)
		return
	}

	if len(releases) == 0 {
		response := map[string]interface{}{
			"update_available": false,
			"latest_version":   "unknown",
			"current_version":  BuildNumber,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	latestVersion := releases[0].TagName

	// Compare with BuildNumber
	// Assuming BuildNumber could be a date or a version tag like "0.0.1"
	// For simplicity, if latestVersion != BuildNumber and not "development", we assume an update is available.
	// If BuildNumber is a date (e.g., 2024-01-01), this comparison might be tricky, but since it's injected,
	// maybe it's best to just say update available if they differ and BuildNumber != "development".

	updateAvailable := false
	if BuildNumber != "development" && BuildNumber != latestVersion {
		updateAvailable = true
	} else if BuildNumber == "development" {
		// During development, let's assume we can always show the update if a release exists.
		updateAvailable = latestVersion != ""
	}

	response := map[string]interface{}{
		"update_available": updateAvailable,
		"latest_version":   latestVersion,
		"current_version":  BuildNumber,
	}
	json.NewEncoder(w).Encode(response)
}

// handleSystemUpdateInstall downloads the latest .deb package and installs it.
func handleSystemUpdateInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "installing"}`))

	updateMu.Lock()
	if updateState == "installing" {
		updateMu.Unlock()
		http.Error(w, "Update already in progress", http.StatusConflict)
		return
	}
	updateState = "installing"
	updateLogs = []string{"Starting update process..."}
	updateMu.Unlock()

	// Run the update process in the background
	go func() {
		defer func() {
			updateMu.Lock()
			if updateState == "installing" {
				updateState = "error"
			}
			updateMu.Unlock()
		}()

		// Give the HTTP response time to be sent
		time.Sleep(2 * time.Second)

		logUpdate("Fetching latest release info from GitHub...")
		// Fetch latest releases info
		req, err := http.NewRequest("GET", "https://api.github.com/repos/webdotpulse/NEMS/releases", nil)
		if err != nil {
			logUpdate("Failed to create request for latest release: %v", err)
			return
		}

		var githubToken string
		db.QueryRow("SELECT github_token FROM site_settings WHERE id = 1").Scan(&githubToken)
		if githubToken != "" {
			req.Header.Set("Authorization", "Bearer "+githubToken)
		}

		req.Header.Set("User-Agent", "NEMS-Update-Installer")
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logUpdate("Failed to fetch latest release: %v", err)
			return
		}
		defer resp.Body.Close()

		var releases []struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			logUpdate("Failed to parse release info: %v", err)
			return
		}

		if len(releases) == 0 {
			logUpdate("No releases found")
			return
		}

		tagName := releases[0].TagName
		version := strings.TrimPrefix(tagName, "v")
		debFile := fmt.Sprintf("nems_%s_arm64.deb", version)
		downloadURL := fmt.Sprintf("https://github.com/webdotpulse/NEMS/releases/download/%s/%s", tagName, debFile)
		targetPath := filepath.Join("/tmp", debFile)

		// Download the file
		logUpdate("Downloading update (%s) to %s...", debFile, targetPath)

		var cmd *exec.Cmd

		if githubToken != "" {
			// If token is provided, add it to curl
			cmd = exec.Command("curl", "-L", "-H", "Authorization: Bearer "+githubToken, "-H", "Accept: application/octet-stream", "-s", "-o", targetPath, downloadURL)
		} else {
			cmd = exec.Command("curl", "-L", "-s", "-o", targetPath, downloadURL)
		}

		if out, err := cmd.CombinedOutput(); err != nil {
			logUpdate("Failed to download update: %v\nOutput:\n%s", err, string(out))
			return
		}
		logUpdate("Download completed successfully.")

		// Create a flag file so the deb postinst script knows an update is happening
		// from the web UI and shouldn't restart the service mid-install.
		if err := os.WriteFile("/tmp/nems_web_update_in_progress", []byte("1"), 0644); err != nil {
			logUpdate("Failed to write update flag file: %v", err)
		}

		// Install the package
		logUpdate("Installing package %s...", debFile)
		cmd = exec.Command("sudo", "dpkg", "-i", targetPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			logUpdate("Failed to install update: %v\nOutput:\n%s", err, string(out))
			return
		}
		logUpdate("Installation successful.")

		updateMu.Lock()
		updateState = "done"
		updateMu.Unlock()

		// Restart the service
		logUpdate("Restarting nems service. The connection will be lost momentarily...")
		cmd = exec.Command("sudo", "systemctl", "restart", "nems.service")
		if out, err := cmd.CombinedOutput(); err != nil {
			logUpdate("Failed to restart service: %v\nOutput:\n%s", err, string(out))
			return
		}
		logUpdate("Update and restart completed successfully.")
	}()
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
		out, err := exec.Command("sudo", "systemctl", "reboot").CombinedOutput()
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

// handleTariffForecast is the HTTP handler for the /api/tariffs/forecast endpoint.
// It retrieves EPEX spot prices for the next 24 hours.
func handleTariffForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	start := now.UTC()
	end := now.Add(24 * time.Hour).UTC()

	var settings models.SiteSettings
	row := db.QueryRow("SELECT contract_type, fixed_price_peak_kwh, fixed_price_off_peak_kwh, fixed_inject_price_kwh, dynamic_markup_kwh, engie_markup_peak, engie_markup_off_peak, engie_markup_super_off_peak, engie_multiplier, engie_base_fee, luminus_markup, luminus_multiplier, luminus_base_fee, eneco_markup, eneco_multiplier, eneco_base_fee, frank_markup, frank_multiplier, frank_base_fee, ecopower_markup, ecopower_multiplier, ecopower_base_fee FROM site_settings WHERE id = 1")
	err = row.Scan(&settings.ContractType, &settings.FixedPricePeakKwh, &settings.FixedPriceOffPeakKwh, &settings.FixedInjectPriceKwh, &settings.DynamicMarkupKwh, &settings.EngieMarkupPeak, &settings.EngieMarkupOffPeak, &settings.EngieMarkupSuperOffPeak, &settings.EngieMultiplier, &settings.EngieBaseFee, &settings.LuminusMarkup, &settings.LuminusMultiplier, &settings.LuminusBaseFee, &settings.EnecoMarkup, &settings.EnecoMultiplier, &settings.EnecoBaseFee, &settings.FrankMarkup, &settings.FrankMultiplier, &settings.FrankBaseFee, &settings.EcopowerMarkup, &settings.EcopowerMultiplier, &settings.EcopowerBaseFee)
	if err != nil {
		if err == sql.ErrNoRows {
			settings.ContractType = "dynamic"
			settings.DynamicMarkupKwh = 0.0
		} else {
			log.Printf("[ERROR] handleTariffForecast: Error fetching site settings: %v", err)
		}
	}

	rows, err := db.Query("SELECT timestamp, price_per_kwh FROM epex_prices WHERE timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC", start, end)
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
		p.PricePerKwh = CalculateEffectivePrice(p.Timestamp, p.PricePerKwh, settings, false)
		prices = append(prices, p)
	}

	if prices == nil {
		prices = []PricePoint{}
	}
	json.NewEncoder(w).Encode(prices)
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

	var settings models.SiteSettings
	row := db.QueryRow("SELECT contract_type, fixed_price_peak_kwh, fixed_price_off_peak_kwh, fixed_inject_price_kwh, dynamic_markup_kwh, engie_markup_peak, engie_markup_off_peak, engie_markup_super_off_peak, engie_multiplier, engie_base_fee, luminus_markup, luminus_multiplier, luminus_base_fee, eneco_markup, eneco_multiplier, eneco_base_fee, frank_markup, frank_multiplier, frank_base_fee, ecopower_markup, ecopower_multiplier, ecopower_base_fee FROM site_settings WHERE id = 1")
	err = row.Scan(&settings.ContractType, &settings.FixedPricePeakKwh, &settings.FixedPriceOffPeakKwh, &settings.FixedInjectPriceKwh, &settings.DynamicMarkupKwh, &settings.EngieMarkupPeak, &settings.EngieMarkupOffPeak, &settings.EngieMarkupSuperOffPeak, &settings.EngieMultiplier, &settings.EngieBaseFee, &settings.LuminusMarkup, &settings.LuminusMultiplier, &settings.LuminusBaseFee, &settings.EnecoMarkup, &settings.EnecoMultiplier, &settings.EnecoBaseFee, &settings.FrankMarkup, &settings.FrankMultiplier, &settings.FrankBaseFee, &settings.EcopowerMarkup, &settings.EcopowerMultiplier, &settings.EcopowerBaseFee)
	if err != nil {
		if err == sql.ErrNoRows {
			settings.ContractType = "dynamic"
			settings.DynamicMarkupKwh = 0.0
		} else {
			log.Printf("[ERROR] handleTariffsToday: Error fetching site settings: %v", err)
		}
	}

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
		p.PricePerKwh = CalculateEffectivePrice(p.Timestamp, p.PricePerKwh, settings, false)
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
		row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, battery_grid_charge_strategy, force_charge_below_euro, force_discharge_above_euro, smart_ev_cheapest_hours, grid_nominal_current_a, grid_system, allowed_grid_import_kw, allowed_grid_export_kw, appliance_turn_on_excess_w, peak_shaving_buffer_w, peak_shaving_rampup_w, timezone, language, address, latitude, longitude, contract_type, fixed_price_peak_kwh, fixed_price_off_peak_kwh, fixed_inject_price_kwh, dynamic_markup_kwh, dynamic_inject_multiplier, engie_markup_peak, engie_markup_off_peak, engie_markup_super_off_peak, engie_multiplier, engie_inject_multiplier, engie_base_fee, luminus_markup, luminus_multiplier, luminus_inject_multiplier, luminus_base_fee, eneco_markup, eneco_multiplier, eneco_inject_multiplier, eneco_base_fee, frank_markup, frank_multiplier, frank_inject_multiplier, frank_base_fee, ecopower_markup, ecopower_multiplier, ecopower_inject_multiplier, ecopower_base_fee, custom_charge_schedule, superdal_optimization_enabled, superdal_target_soc, home_base_load_w, github_token FROM site_settings WHERE id = 1")
		var settings models.SiteSettings
		err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.BatteryGridChargeStrategy, &settings.ForceChargeBelowEuro, &settings.ForceDischargeAboveEuro, &settings.SmartEvCheapestHours, &settings.GridNominalCurrentA, &settings.GridSystem, &settings.AllowedGridImportKw, &settings.AllowedGridExportKw, &settings.ApplianceTurnOnExcessW, &settings.PeakShavingBufferW, &settings.PeakShavingRampupW, &settings.Timezone, &settings.Language, &settings.Address, &settings.Latitude, &settings.Longitude, &settings.ContractType, &settings.FixedPricePeakKwh, &settings.FixedPriceOffPeakKwh, &settings.FixedInjectPriceKwh, &settings.DynamicMarkupKwh, &settings.DynamicInjectMultiplier, &settings.EngieMarkupPeak, &settings.EngieMarkupOffPeak, &settings.EngieMarkupSuperOffPeak, &settings.EngieMultiplier, &settings.EngieInjectMultiplier, &settings.EngieBaseFee, &settings.LuminusMarkup, &settings.LuminusMultiplier, &settings.LuminusInjectMultiplier, &settings.LuminusBaseFee, &settings.EnecoMarkup, &settings.EnecoMultiplier, &settings.EnecoInjectMultiplier, &settings.EnecoBaseFee, &settings.FrankMarkup, &settings.FrankMultiplier, &settings.FrankInjectMultiplier, &settings.FrankBaseFee, &settings.EcopowerMarkup, &settings.EcopowerMultiplier, &settings.EcopowerInjectMultiplier, &settings.EcopowerBaseFee, &settings.CustomChargeSchedule, &settings.SuperdalOptimizationEnabled, &settings.SuperdalTargetSoc, &settings.HomeBaseLoadW, &settings.GithubToken)
		if err != nil {
			if err == sql.ErrNoRows {
				// Fallback
				settings = models.SiteSettings{
					StrategyMode:                "eco",
					CapacityPeakLimitKw:         2.5,
					ActiveInverterCurtailment:   false,
					BatteryGridChargeStrategy:   "price_only",
					ForceChargeBelowEuro:        0.0,
					ForceDischargeAboveEuro:     999.0,
					SmartEvCheapestHours:        0,
					GridNominalCurrentA:         25.0,
					GridSystem:                  "single_phase_230v",
					AllowedGridImportKw:         0.0,
					AllowedGridExportKw:         0.0,
					ApplianceTurnOnExcessW:      0.0,
					PeakShavingBufferW:          200.0,
					PeakShavingRampupW:          500.0,
					Timezone:                    "Europe/Brussels",
					Language:                    "EN",
					Address:                     "",
					Latitude:                    50.8503,
					Longitude:                   4.3517,
					ContractType:                "dynamic",
					FixedPricePeakKwh:           0.35,
					FixedPriceOffPeakKwh:        0.30,
					FixedInjectPriceKwh:         0.05,
					DynamicMarkupKwh:            0.15,
					EngieMarkupPeak:             0.15,
					EngieMarkupOffPeak:          0.15,
					EngieMarkupSuperOffPeak:     0.15,
					EngieMultiplier:             0.1448,
					EngieBaseFee:                0.0,
					CustomChargeSchedule:        "[]",
					SuperdalOptimizationEnabled: false,
					SuperdalTargetSoc:           100.0,
					HomeBaseLoadW:               300.0,
					GithubToken:                 "",
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

		if settings.BatteryGridChargeStrategy == "" {
			settings.BatteryGridChargeStrategy = "price_only"
		}

		if settings.CustomChargeSchedule == "" {
			settings.CustomChargeSchedule = "[]"
		}

		_, err := db.Exec("UPDATE site_settings SET strategy_mode = ?, capacity_peak_limit_kw = ?, active_inverter_curtailment = ?, battery_grid_charge_strategy = ?, force_charge_below_euro = ?, force_discharge_above_euro = ?, smart_ev_cheapest_hours = ?, grid_nominal_current_a = ?, grid_system = ?, allowed_grid_import_kw = ?, allowed_grid_export_kw = ?, appliance_turn_on_excess_w = ?, peak_shaving_buffer_w = ?, peak_shaving_rampup_w = ?, timezone = ?, language = ?, address = ?, latitude = ?, longitude = ?, contract_type = ?, fixed_price_peak_kwh = ?, fixed_price_off_peak_kwh = ?, fixed_inject_price_kwh = ?, dynamic_markup_kwh = ?, dynamic_inject_multiplier = ?, engie_markup_peak = ?, engie_markup_off_peak = ?, engie_markup_super_off_peak = ?, engie_multiplier = ?, engie_inject_multiplier = ?, engie_base_fee = ?, luminus_markup = ?, luminus_multiplier = ?, luminus_inject_multiplier = ?, luminus_base_fee = ?, eneco_markup = ?, eneco_multiplier = ?, eneco_inject_multiplier = ?, eneco_base_fee = ?, frank_markup = ?, frank_multiplier = ?, frank_inject_multiplier = ?, frank_base_fee = ?, ecopower_markup = ?, ecopower_multiplier = ?, ecopower_inject_multiplier = ?, ecopower_base_fee = ?, custom_charge_schedule = ?, superdal_optimization_enabled = ?, superdal_target_soc = ?, home_base_load_w = ?, github_token = ? WHERE id = 1",
			settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment, settings.BatteryGridChargeStrategy, settings.ForceChargeBelowEuro, settings.ForceDischargeAboveEuro, settings.SmartEvCheapestHours, settings.GridNominalCurrentA, settings.GridSystem, settings.AllowedGridImportKw, settings.AllowedGridExportKw, settings.ApplianceTurnOnExcessW, settings.PeakShavingBufferW, settings.PeakShavingRampupW, settings.Timezone, settings.Language, settings.Address, settings.Latitude, settings.Longitude, settings.ContractType, settings.FixedPricePeakKwh, settings.FixedPriceOffPeakKwh, settings.FixedInjectPriceKwh, settings.DynamicMarkupKwh, settings.DynamicInjectMultiplier, settings.EngieMarkupPeak, settings.EngieMarkupOffPeak, settings.EngieMarkupSuperOffPeak, settings.EngieMultiplier, settings.EngieInjectMultiplier, settings.EngieBaseFee, settings.LuminusMarkup, settings.LuminusMultiplier, settings.LuminusInjectMultiplier, settings.LuminusBaseFee, settings.EnecoMarkup, settings.EnecoMultiplier, settings.EnecoInjectMultiplier, settings.EnecoBaseFee, settings.FrankMarkup, settings.FrankMultiplier, settings.FrankInjectMultiplier, settings.FrankBaseFee, settings.EcopowerMarkup, settings.EcopowerMultiplier, settings.EcopowerInjectMultiplier, settings.EcopowerBaseFee, settings.CustomChargeSchedule, settings.SuperdalOptimizationEnabled, settings.SuperdalTargetSoc, settings.HomeBaseLoadW, settings.GithubToken)
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
		rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, inverter_rated_power_kw, charge_mode, battery_mode, ocpp_proxy_url, poll_interval FROM devices")
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
			var pollInterval sql.NullInt64
			if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password, &d.HasGridMeter, &d.HasBattery, &d.BatteryCapacity, &d.InverterRatedPowerKw, &chargeMode, &batteryMode, &proxyUrl, &pollInterval); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if pollInterval.Valid {
				d.PollInterval = int(pollInterval.Int64)
			} else {
				d.PollInterval = 5
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
		if d.PollInterval == 0 {
			d.PollInterval = 5
		}

		result, err := db.Exec("INSERT INTO devices (name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, inverter_rated_power_kw, charge_mode, battery_mode, ocpp_proxy_url, poll_interval) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.InverterRatedPowerKw, d.ChargeMode, d.BatteryMode, d.OcppProxyUrl, d.PollInterval)
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
		if d.PollInterval == 0 {
			d.PollInterval = 5
		}

		_, err = db.Exec("UPDATE devices SET name = ?, template = ?, host = ?, port = ?, modbus_id = ?, username = ?, password = ?, has_grid_meter = ?, has_battery = ?, battery_capacity = ?, inverter_rated_power_kw = ?, charge_mode = ?, battery_mode = ?, ocpp_proxy_url = ?, poll_interval = ? WHERE id = ?",
			d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.InverterRatedPowerKw, d.ChargeMode, d.BatteryMode, d.OcppProxyUrl, d.PollInterval, id)
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

type SolarForecastPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	EstimatedPowerW float64   `json:"estimated_power_w"`
}

// handleSolarForecast is the HTTP handler for the /api/solar/forecast endpoint.
// It retrieves the solar forecast based on the user's configured latitude and longitude
// using the free Open-Meteo API.
func handleSolarForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get latitude and longitude from settings
	var settings models.SiteSettings
	row := db.QueryRow("SELECT latitude, longitude FROM site_settings WHERE id = 1")
	err := row.Scan(&settings.Latitude, &settings.Longitude)
	if err != nil {
		log.Printf("[ERROR] handleSolarForecast: Error fetching site settings: %v", err)
		http.Error(w, "Could not retrieve site settings", http.StatusInternalServerError)
		return
	}

	if settings.Latitude == 0 && settings.Longitude == 0 {
		http.Error(w, "Location (latitude/longitude) is not configured in settings.", http.StatusBadRequest)
		return
	}

	// Fetch from Open-Meteo API
	// Example URL: https://api.open-meteo.com/v1/forecast?latitude=52.52&longitude=13.41&hourly=direct_radiation,diffuse_radiation
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=direct_radiation,diffuse_radiation&forecast_days=2", settings.Latitude, settings.Longitude)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ERROR] handleSolarForecast: Error calling Open-Meteo API: %v", err)
		http.Error(w, "Failed to fetch solar forecast from external API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] handleSolarForecast: Open-Meteo API returned status: %d", resp.StatusCode)
		http.Error(w, "External API error", http.StatusInternalServerError)
		return
	}

	var meteoData struct {
		Hourly struct {
			Time             []string  `json:"time"`
			DirectRadiation  []float64 `json:"direct_radiation"`
			DiffuseRadiation []float64 `json:"diffuse_radiation"`
		} `json:"hourly"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meteoData); err != nil {
		log.Printf("[ERROR] handleSolarForecast: Error decoding Open-Meteo API response: %v", err)
		http.Error(w, "Failed to decode forecast data", http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation("Europe/Amsterdam") // Default fallback if needed, but time string from meteo is iso8601
	if err != nil {
		loc = time.UTC
	}

	var forecast []SolarForecastPoint
	now := time.Now().In(loc)
	start := now.Truncate(time.Hour) // start from the current hour
	end := start.Add(24 * time.Hour) // next 24 hours

	for i, tStr := range meteoData.Hourly.Time {
		// Open-meteo returns time in ISO8601 like "2023-10-25T00:00"
		t, err := time.Parse("2006-01-02T15:04", tStr)
		if err != nil {
			continue
		}

		// Adjust timezone to local if necessary, open-meteo usually returns local time or UTC based on params
		// But let's assume it's UTC or we can just use the parsed time.
		// Open-Meteo defaults to UTC if no timezone is specified.
		tUTC := t.UTC()

		if tUTC.Before(start.UTC()) {
			continue
		}
		if tUTC.After(end.UTC()) {
			break
		}

		// Calculate estimated power.
		// Direct + Diffuse radiation (W/m²). We just use this raw value as a proxy for estimated generation.
		// A more complex model would multiply by area and efficiency.
		direct := 0.0
		diffuse := 0.0
		if i < len(meteoData.Hourly.DirectRadiation) {
			direct = meteoData.Hourly.DirectRadiation[i]
		}
		if i < len(meteoData.Hourly.DiffuseRadiation) {
			diffuse = meteoData.Hourly.DiffuseRadiation[i]
		}

		estimatedPower := direct + diffuse

		forecast = append(forecast, SolarForecastPoint{
			Timestamp:       tUTC.In(loc),
			EstimatedPowerW: estimatedPower,
		})
	}

	if forecast == nil {
		forecast = []SolarForecastPoint{}
	}

	json.NewEncoder(w).Encode(forecast)
}
