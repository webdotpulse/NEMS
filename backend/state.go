package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type SiteState struct {
	GridPowerW      *float64         `json:"grid_power_w"`
	SolarPowerW     *float64         `json:"solar_power_w"`
	BatteryPowerW   *float64         `json:"battery_power_w"`
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

	whereClause := ""
	switch node {
	case "grid":
		whereClause = "template = 'huawei_dongle'"
	case "solar":
		whereClause = "template = 'huawei_inverter' AND name NOT LIKE '%battery%'"
	case "battery":
		whereClause = "template = 'huawei_inverter' AND name LIKE '%battery%'"
	case "ev_charger":
		whereClause = "template = 'raedian_charger'"
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

	// 2. Determine time constraint and interval grouping
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

	query := fmt.Sprintf(`
		SELECT
			%s AS ts,
			AVG(power_w) as avg_power
		FROM measurements
		WHERE device_id IN (%s) AND %s
		GROUP BY ts
		ORDER BY ts ASC
	`, groupByClause, inClause, timeConstraint)

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
