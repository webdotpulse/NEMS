package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SiteState represents the aggregated real-time power metrics of the entire system.
// Pointers to float64 are used so that unconfigured or offline devices can be serialized
// as `null` in JSON rather than defaulting to `0`.
type SiteState struct {
	GridPowerW            *float64       `json:"grid_power_w"`
	SolarPowerW           *float64       `json:"solar_power_w"`
	BatteryPowerW         *float64       `json:"battery_power_w"`
	BatterySoc            *float64       `json:"battery_soc"`
	TotalLoadW            *float64       `json:"total_load_w"`
	EvChargerPowerW       *float64       `json:"ev_charger_power_w"`
	ProjectedQuarterPeakW *float64       `json:"projected_quarter_peak_w"`
	DeviceHealth          map[int]string `json:"device_health"`
}

// StateDispatcher manages Server-Sent Events (SSE) connections for clients
// listening to real-time site state updates.
type StateDispatcher struct {
	clients map[chan SiteState]bool
	mu      sync.Mutex
}

// GlobalStateDispatcher is the singleton instance handling all active SSE clients.
var GlobalStateDispatcher = &StateDispatcher{
	clients: make(map[chan SiteState]bool),
}

// AddClient creates a new buffered channel for a client and registers it with the dispatcher.
func (d *StateDispatcher) AddClient() chan SiteState {
	d.mu.Lock()
	defer d.mu.Unlock()
	ch := make(chan SiteState, 1) // Buffer to avoid blocking the broadcaster
	d.clients[ch] = true
	return ch
}

// RemoveClient unregisters a client's channel and closes it.
func (d *StateDispatcher) RemoveClient(ch chan SiteState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.clients[ch]; ok {
		delete(d.clients, ch)
		close(ch)
	}
}

// Broadcast sends the current SiteState to all registered SSE clients.
// If a client's channel is full, the message is dropped to prevent blocking the polling loop.
func (d *StateDispatcher) Broadcast(state SiteState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for ch := range d.clients {
		select {
		case ch <- state:
		default:
			// Client channel full, skip to avoid blocking the broadcaster
		}
	}
}

// handleLiveStream is the HTTP handler for the /api/live SSE endpoint.
func handleLiveStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := GlobalStateDispatcher.AddClient()
	defer GlobalStateDispatcher.RemoveClient(clientChan)

	// Send an initial connected message
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			log.Println("SSE client disconnected")
			return
		case state := <-clientChan:
			data, err := json.Marshal(state)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// DailyAggregates holds the calculated energy summaries for the current day.
type DailyAggregates struct {
	GridImportKwh         float64 `json:"grid_import_kwh"`
	GridExportKwh         float64 `json:"grid_export_kwh"`
	SolarYieldKwh         float64 `json:"solar_yield_kwh"`
	BatteryChargeKwh      float64 `json:"battery_charge_kwh"`
	BatteryChargeSolarKwh float64 `json:"battery_charge_solar_kwh"`
	BatteryChargeGridKwh  float64 `json:"battery_charge_grid_kwh"`
	BatteryDischargeKwh   float64 `json:"battery_discharge_kwh"`
	HouseConsumptionKwh   float64 `json:"house_consumption_kwh"`
}

// handleDailyAggregates computes today's energy statistics dynamically via a CTE query.
func handleDailyAggregates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		// Fallback to UTC if tzdata is missing on the system
		loc = time.UTC
	}
	now := time.Now().In(loc)

	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()

	// Parse optional date parameter
	dateParam := r.URL.Query().Get("date")
	if dateParam != "" {
		parsedDate, err := time.ParseInLocation("2006-01-02", dateParam, loc)
		if err == nil {
			startOfToday = parsedDate.UTC()
		}
	}
	endOfDay := startOfToday.Add(24 * time.Hour)

	// Group by minute to do the charge source calculation properly.
	// This determines if a battery was charged from excess solar or from the grid.
	query := `
		WITH minute_aggs AS (
			SELECT
				strftime('%Y-%m-%dT%H:%M:00Z', m.timestamp) as ts,
				SUM(CASE WHEN (d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (d.template = 'huawei_inverter' AND d.has_grid_meter = 1)) AND m.power_w > 0 THEN (m.power_w / 60000.0) ELSE 0 END) as grid_import,
				SUM(CASE WHEN (d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (d.template = 'huawei_inverter' AND d.has_grid_meter = 1)) AND m.power_w < 0 THEN ABS(m.power_w / 60000.0) ELSE 0 END) as grid_export,
				SUM(CASE WHEN d.template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'demo_inverter') AND d.name NOT LIKE '%battery%' AND m.power_w > 0 THEN (m.power_w / 60000.0) ELSE 0 END) as solar_yield,
				SUM(CASE WHEN ((d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%battery%') OR d.template = 'demo_battery') AND m.power_w < 0 THEN ABS(m.power_w / 60000.0) ELSE 0 END) as battery_charge,
				SUM(CASE WHEN ((d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%battery%') OR d.template = 'demo_battery') AND m.power_w > 0 THEN (m.power_w / 60000.0) ELSE 0 END) as battery_discharge
			FROM measurements m
			JOIN devices d ON m.device_id = d.id
			WHERE m.timestamp >= ? AND m.timestamp < ?
			GROUP BY ts
		)
		SELECT
			SUM(grid_import),
			SUM(grid_export),
			SUM(solar_yield),
			SUM(battery_charge),
			SUM(battery_discharge),
			SUM(CASE
				WHEN battery_charge > 0 AND solar_yield > 0
				THEN MIN(battery_charge, solar_yield)
				ELSE 0
			END) as battery_charge_solar,
			SUM(CASE
				WHEN battery_charge > 0 AND solar_yield > 0
				THEN MAX(0, battery_charge - solar_yield)
				WHEN battery_charge > 0 AND solar_yield <= 0
				THEN battery_charge
				ELSE 0
			END) as battery_charge_grid
		FROM minute_aggs
	`

	row := db.QueryRow(query, startOfToday.Format("2006-01-02 15:04:05"), endOfDay.Format("2006-01-02 15:04:05"))

	var agg DailyAggregates
	var gImport, gExport, sYield, bCharge, bDischarge, bChargeSolar, bChargeGrid *float64

	err = row.Scan(&gImport, &gExport, &sYield, &bCharge, &bDischarge, &bChargeSolar, &bChargeGrid)
	if err != nil {
		if err == sql.ErrNoRows {
			// No measurements today, return empty struct
			json.NewEncoder(w).Encode(agg)
			return
		}
		// Some other db error occurred
		log.Printf("Error scanning daily aggregates: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Safe pointer dereferencing for aggregate sum columns (which can be NULL if CTE result is empty)
	if gImport != nil { agg.GridImportKwh = *gImport }
	if gExport != nil { agg.GridExportKwh = *gExport }
	if sYield != nil { agg.SolarYieldKwh = *sYield }
	if bCharge != nil { agg.BatteryChargeKwh = *bCharge }
	if bDischarge != nil { agg.BatteryDischargeKwh = *bDischarge }
	if bChargeSolar != nil { agg.BatteryChargeSolarKwh = *bChargeSolar }
	if bChargeGrid != nil { agg.BatteryChargeGridKwh = *bChargeGrid }

	// Calculate net house consumption
	agg.HouseConsumptionKwh = agg.GridImportKwh + agg.SolarYieldKwh + agg.BatteryDischargeKwh - agg.GridExportKwh - agg.BatteryChargeKwh
	if agg.HouseConsumptionKwh < 0 {
		agg.HouseConsumptionKwh = 0
	}

	json.NewEncoder(w).Encode(agg)
}

// HistoryDataPoint represents a single data point on the frontend charts.
type HistoryDataPoint struct {
	Timestamp string  `json:"timestamp"`
	PowerW    float64 `json:"power_w"`
}

// handleHistory aggregates historical power measurements over varying intervals (5m to 1h)
// based on the requested time range and node (e.g. grid, solar, home).
func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	node := r.URL.Query().Get("node")
	timeRange := r.URL.Query().Get("range")

	if node == "" {
		node = "grid"
	}
	if timeRange == "" {
		timeRange = "today"
	}

	// 1. Map node to device type criteria
	// For device criteria matching memory:
	// huawei_inverter (for solar and battery, battery if name contains 'battery')
	// huawei_dongle (for grid)
	// raedian_charger (for EV charger)

	// Determine time constraint and interval grouping
	// today: >= datetime('now', 'start of day'), 5-minute interval
	// 24h: >= datetime('now', '-24 hours'), 5-minute interval
	// 7d: >= datetime('now', '-7 days'), 1-hour interval
	// 30d: >= datetime('now', '-30 days'), 1-hour interval

	timeConstraint := ""
	groupByClause := ""

	switch timeRange {
	case "today":
		timeConstraint = "timestamp >= datetime('now', 'start of day')"
		groupByClause = "strftime('%Y-%m-%dT%H:', timestamp) || printf('%02d', (CAST(strftime('%M', timestamp) AS INTEGER) / 5) * 5) || ':00Z'"
	case "24h":
		timeConstraint = "timestamp >= datetime('now', '-24 hours')"
		groupByClause = "strftime('%Y-%m-%dT%H:', timestamp) || printf('%02d', (CAST(strftime('%M', timestamp) AS INTEGER) / 5) * 5) || ':00Z'"
	case "7d":
		timeConstraint = "timestamp >= datetime('now', '-7 days')"
		groupByClause = "strftime('%Y-%m-%dT%H:00:00Z', timestamp)"
	case "30d":
		timeConstraint = "timestamp >= datetime('now', '-30 days')"
		groupByClause = "strftime('%Y-%m-%dT%H:00:00Z', timestamp)"
	default:
		http.Error(w, "Invalid range", http.StatusBadRequest)
		return
	}

	var query string

	if node == "home" {
		query = fmt.Sprintf(`
			SELECT
				%s AS ts,
				SUM(
					CASE
						WHEN d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (d.template = 'huawei_inverter' AND d.has_grid_meter = 1) THEN m.power_w
						WHEN d.template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'demo_inverter') AND d.name NOT LIKE '%%battery%%' THEN m.power_w
						WHEN d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%%battery%%' THEN m.power_w
						WHEN d.template IN ('raedian_charger', 'alfen_charger', 'bender_charger', 'phoenix_charger', 'easee_charger', 'peblar_charger', 'demo_charger') THEN -m.power_w
						ELSE 0
					END
				) / COUNT(DISTINCT d.id) as avg_power
			FROM measurements m
			JOIN devices d ON CAST(m.device_id AS INTEGER) = d.id
			WHERE %s
			GROUP BY ts
			ORDER BY ts ASC
		`, groupByClause, timeConstraint)
	} else {
		whereClause := ""
		switch node {
		case "grid":
			whereClause = "(template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (template = 'huawei_inverter' AND has_grid_meter = 1))"
		case "solar":
			whereClause = "(template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'demo_inverter') AND name NOT LIKE '%battery%')"
		case "battery":
			whereClause = "(template IN ('huawei_inverter', 'demo_inverter') AND name LIKE '%battery%') OR template = 'demo_battery'"
		case "ev_charger":
			whereClause = "(template IN ('raedian_charger', 'alfen_charger', 'bender_charger', 'phoenix_charger', 'easee_charger', 'peblar_charger', 'demo_charger'))"
		default:
			http.Error(w, "Invalid node", http.StatusBadRequest)
			return
		}

		// Fetch applicable device IDs
		rows, err := db.Query(fmt.Sprintf("SELECT id FROM devices WHERE %s", whereClause))
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var deviceIDs []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				deviceIDs = append(deviceIDs, id)
			}
		}

		if len(deviceIDs) == 0 {
			// Return empty array if no devices matched
			json.NewEncoder(w).Encode([]HistoryDataPoint{})
			return
		}

		// Build IN clause for device_id
		inClause := ""
		for i, id := range deviceIDs {
			if i > 0 {
				inClause += ","
			}
			inClause += fmt.Sprintf("'%d'", id)
		}

		query = fmt.Sprintf(`
			SELECT
				%s AS ts,
				AVG(power_w) as avg_power
			FROM measurements
			WHERE device_id IN (%s) AND %s
			GROUP BY ts
			ORDER BY ts ASC
		`, groupByClause, inClause, timeConstraint)
	}

	measurementsRows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer measurementsRows.Close()

	var data []HistoryDataPoint
	for measurementsRows.Next() {
		var dp HistoryDataPoint
		if err := measurementsRows.Scan(&dp.Timestamp, &dp.PowerW); err == nil {
			data = append(data, dp)
		}
	}

	if data == nil {
		data = []HistoryDataPoint{}
	}

	json.NewEncoder(w).Encode(data)
}
