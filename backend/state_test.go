package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

// oldBroadcast simulates the old broadcasting logic where we range over clients
// and push the raw struct into the channel, forcing each client handler to marshal it
func oldBroadcast(state SiteState, numClients int) {
	// Simulate what handleLiveStream used to do per client
	for i := 0; i < numClients; i++ {
		data, err := json.Marshal(state)
		if err != nil {
			continue
		}
		_ = fmt.Sprintf("data: %s\n\n", string(data)) // Simulate fmt.Fprintf allocation
	}
}

// newBroadcast simulates the new optimized broadcasting logic where we marshal once
// and pre-build the []byte, simulating pushing the byte slice to multiple clients.
func newBroadcast(state SiteState, numClients int) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	msg := make([]byte, 0, len("data: ")+len(data)+len("\n\n"))
	msg = append(msg, "data: "...)
	msg = append(msg, data...)
	msg = append(msg, "\n\n"...)

	// Simulate what handleLiveStream does now per client (just a write of the existing slice)
	for i := 0; i < numClients; i++ {
		_ = msg
	}
}

func BenchmarkSSEBroadcasting(b *testing.B) {
	vGrid, vSolar, vBatt, vLoad, vSoc := 1500.0, 3000.0, -1000.0, 2500.0, 50.0

	state := SiteState{
		GridPowerW:    &vGrid,
		SolarPowerW:   &vSolar,
		BatteryPowerW: &vBatt,
		TotalLoadW:    &vLoad,
		BatterySoc:    &vSoc,
		DeviceHealth:  map[int]string{1: "ok", 2: "ok"},
	}

	numClients := 10

	b.Run("Old_MultiMarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			oldBroadcast(state, numClients)
		}
	})

	b.Run("New_SingleMarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newBroadcast(state, numClients)
		}
	})
}
