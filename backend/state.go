package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type SiteState struct {
	GridPowerW      *float64         `json:"grid_power_w"`
	SolarPowerW     *float64         `json:"solar_power_w"`
	BatteryPowerW   *float64         `json:"battery_power_w"`
	BatterySoc      *float64         `json:"battery_soc"`
	TotalLoadW      *float64         `json:"total_load_w"`
	EvChargerPowerW *float64         `json:"ev_charger_power_w"`
	DeviceHealth    map[int]string   `json:"device_health"`
}

type StateDispatcher struct {
	clients map[chan SiteState]bool
	mu      sync.Mutex
}

var GlobalStateDispatcher = &StateDispatcher{
	clients: make(map[chan SiteState]bool),
}

func (d *StateDispatcher) AddClient() chan SiteState {
	d.mu.Lock()
	defer d.mu.Unlock()
	ch := make(chan SiteState, 1) // buffer to avoid blocking
	d.clients[ch] = true
	return ch
}

func (d *StateDispatcher) RemoveClient(ch chan SiteState) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.clients[ch]; ok {
		delete(d.clients, ch)
		close(ch)
	}
}

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

type DailyAggregates struct {
	GridImportKwh           float64 `json:"grid_import_kwh"`
	GridExportKwh           float64 `json:"grid_export_kwh"`
	SolarYieldKwh           float64 `json:"solar_yield_kwh"`
	BatteryChargeKwh        float64 `json:"battery_charge_kwh"`
	BatteryChargeSolarKwh   float64 `json:"battery_charge_solar_kwh"`
	BatteryChargeGridKwh    float64 `json:"battery_charge_grid_kwh"`
	BatteryDischargeKwh     float64 `json:"battery_discharge_kwh"`
	HouseConsumptionKwh     float64 `json:"house_consumption_kwh"`
}

func handleDailyAggregates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()

	// 1. Group by minute to do the charge source calculation properly.
	query := `
		WITH minute_aggs AS (
			SELECT
				strftime('%Y-%m-%dT%H:%M:00Z', m.timestamp) as ts,
				SUM(CASE WHEN d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter') AND m.power_w > 0 THEN m.energy_kwh ELSE 0 END) as grid_import,
				SUM(CASE WHEN d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter') AND m.power_w < 0 THEN ABS(m.energy_kwh) ELSE 0 END) as grid_export,
				SUM(CASE WHEN d.template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'demo_inverter') AND d.name NOT LIKE '%battery%' AND m.power_w > 0 THEN m.energy_kwh ELSE 0 END) as solar_yield,
				SUM(CASE WHEN (d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%battery%' OR d.template = 'demo_battery') AND m.power_w < 0 THEN ABS(m.energy_kwh) ELSE 0 END) as battery_charge,
				SUM(CASE WHEN (d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%battery%' OR d.template = 'demo_battery') AND m.power_w > 0 THEN m.energy_kwh ELSE 0 END) as battery_discharge
			FROM measurements m
			JOIN devices d ON m.device_id = d.id
			WHERE m.timestamp >= ?
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

	row := db.QueryRow(query, startOfToday.Format("2006-01-02 15:04:05"))

	var agg DailyAggregates
	var gImport, gExport, sYield, bCharge, bDischarge, bChargeSolar, bChargeGrid *float64

	err = row.Scan(&gImport, &gExport, &sYield, &bCharge, &bDischarge, &bChargeSolar, &bChargeGrid)
	if err == nil {
		if gImport != nil { agg.GridImportKwh = *gImport }
		if gExport != nil { agg.GridExportKwh = *gExport }
		if sYield != nil { agg.SolarYieldKwh = *sYield }
		if bCharge != nil { agg.BatteryChargeKwh = *bCharge }
		if bDischarge != nil { agg.BatteryDischargeKwh = *bDischarge }
		if bChargeSolar != nil { agg.BatteryChargeSolarKwh = *bChargeSolar }
		if bChargeGrid != nil { agg.BatteryChargeGridKwh = *bChargeGrid }

		agg.HouseConsumptionKwh = agg.GridImportKwh + agg.SolarYieldKwh + agg.BatteryDischargeKwh - agg.GridExportKwh - agg.BatteryChargeKwh
		if agg.HouseConsumptionKwh < 0 {
			agg.HouseConsumptionKwh = 0
		}
	} else {
		log.Printf("Error calculating daily aggregates: %v", err)
	}

	json.NewEncoder(w).Encode(agg)
}

type HistoryDataPoint struct {
	Timestamp string  `json:"timestamp"`
	PowerW    float64 `json:"power_w"`
}

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
						WHEN d.template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter') THEN m.power_w
						WHEN d.template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'demo_inverter') AND d.name NOT LIKE '%%battery%%' THEN m.power_w
						WHEN d.template IN ('huawei_inverter', 'demo_inverter') AND d.name LIKE '%%battery%%' THEN m.power_w
						WHEN d.template IN ('raedian_charger', 'alfen_charger', 'bender_charger', 'phoenix_charger', 'easee_charger', 'peblar_charger', 'demo_charger') THEN -m.power_w
						ELSE 0
					END
				) / COUNT(DISTINCT d.id) as avg_power
			FROM measurements m
			JOIN devices d ON m.device_id = d.id
			WHERE %s
			GROUP BY ts
			ORDER BY ts ASC
		`, groupByClause, timeConstraint)
	} else {
		whereClause := ""
		switch node {
		case "grid":
			whereClause = "(template IN ('huawei_dongle', 'demo_dongle', 'homewizard_meter'))"
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
			inClause += fmt.Sprintf("%d", id)
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
