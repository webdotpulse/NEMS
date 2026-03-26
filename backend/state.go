package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type SiteState struct {
	GridPowerW      float64 `json:"grid_power_w"`
	SolarPowerW     float64 `json:"solar_power_w"`
	BatteryPowerW   float64 `json:"battery_power_w"`
	TotalLoadW      float64 `json:"total_load_w"`
	EvChargerPowerW float64 `json:"ev_charger_power_w"`
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
