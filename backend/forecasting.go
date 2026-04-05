package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"nems/internal/models"
)

// DynamicBatteryForecast represents the projected optimal plan.
// It maps the hour (Unix timestamp for start of hour) to a string action ("charge", "discharge", "hold").
type DynamicBatteryForecast struct {
	HourlyActions map[int64]string
}

// CalculateDynamicBatteryForecast computes the optimal battery arbitrage strategy for the coming 24 hours.
func CalculateDynamicBatteryForecast(settings models.SiteSettings) DynamicBatteryForecast {
	forecast := DynamicBatteryForecast{
		HourlyActions: make(map[int64]string),
	}

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	startOfHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc).UTC()
	endOfForecast := startOfHour.Add(24 * time.Hour)

	// 1. Fetch EPEX Prices for next 24 hours
	rows, err := db.Query("SELECT timestamp, price_per_kwh FROM epex_prices WHERE timestamp >= ? AND timestamp < ? ORDER BY timestamp ASC", startOfHour, endOfForecast)
	if err != nil {
		log.Printf("[ERROR] Forecasting: Error fetching EPEX prices: %v", err)
		return forecast
	}
	defer rows.Close()

	var prices []PricePoint
	for rows.Next() {
		var p PricePoint
		var ts time.Time
		if err := rows.Scan(&ts, &p.PricePerKwh); err == nil {
			p.Timestamp = ts.UTC()
			p.PricePerKwh = CalculateEffectivePrice(p.Timestamp, p.PricePerKwh, settings, false)
			prices = append(prices, p)
		}
	}

	if len(prices) == 0 {
		return forecast
	}

	// 2. Fetch Solar Forecast
	solarData := fetchSolarForecast(settings.Latitude, settings.Longitude, startOfHour, endOfForecast)

	// Map solar data by hour
	solarByHour := make(map[int64]float64)
	for _, p := range solarData {
		solarByHour[p.Timestamp.UTC().Unix()] = p.EstimatedPowerW
	}

	// 3. Compute Net Load per hour
	// NetLoad_h = HomeBaseLoadW - SolarForecast_h (in W)
	type HourData struct {
		Timestamp time.Time
		Price     float64
		NetLoadW  float64
	}

	var hourlyData []HourData
	for _, p := range prices {
		unixTs := p.Timestamp.Unix()
		solarEst := solarByHour[unixTs]
		netLoad := settings.HomeBaseLoadW - solarEst
		hourlyData = append(hourlyData, HourData{
			Timestamp: p.Timestamp,
			Price:     p.PricePerKwh,
			NetLoadW:  netLoad,
		})
	}

	if len(hourlyData) == 0 {
		return forecast
	}

	// 4. Identify the top 4 most expensive hours
	sortedByPriceDesc := make([]HourData, len(hourlyData))
	copy(sortedByPriceDesc, hourlyData)
	sort.Slice(sortedByPriceDesc, func(i, j int) bool {
		return sortedByPriceDesc[i].Price > sortedByPriceDesc[j].Price
	})

	expensiveHoursCount := 4
	if len(sortedByPriceDesc) < expensiveHoursCount {
		expensiveHoursCount = len(sortedByPriceDesc)
	}

	targetEnergyWh := 0.0
	expensiveSlots := make(map[int64]bool)

	for i := 0; i < expensiveHoursCount; i++ {
		hd := sortedByPriceDesc[i]
		if hd.NetLoadW > 0 {
			targetEnergyWh += hd.NetLoadW * 1.0 // 1 hour duration
		}
		expensiveSlots[hd.Timestamp.Unix()] = true
	}

	// Fetch Total Battery Capacity from DB to cap targetEnergyWh
	totalBatteryCapacityWh := 0.0
	capRow, err := db.Query("SELECT battery_capacity FROM devices WHERE has_battery = 1")
	if err == nil {
		for capRow.Next() {
			var capKwh float64
			if err := capRow.Scan(&capKwh); err == nil {
				totalBatteryCapacityWh += capKwh * 1000.0
			}
		}
		capRow.Close()
	}

	if totalBatteryCapacityWh == 0 {
		return forecast // No batteries configured
	}

	if targetEnergyWh > totalBatteryCapacityWh {
		targetEnergyWh = totalBatteryCapacityWh
	}

	// If no energy is needed to cover net load, do not charge
	if targetEnergyWh <= 0 {
		return forecast
	}

	// 5. Identify the cheapest hours to cover targetEnergyWh
	sortedByPriceAsc := make([]HourData, len(hourlyData))
	copy(sortedByPriceAsc, hourlyData)
	sort.Slice(sortedByPriceAsc, func(i, j int) bool {
		return sortedByPriceAsc[i].Price < sortedByPriceAsc[j].Price
	})

	energyToBuyWh := targetEnergyWh
	chargeSlots := make(map[int64]bool)

	for _, hd := range sortedByPriceAsc {
		if energyToBuyWh <= 0 {
			break
		}
		// Do not charge during an expensive slot
		if expensiveSlots[hd.Timestamp.Unix()] {
			continue
		}

		chargeSlots[hd.Timestamp.Unix()] = true

		// In a single hour slot, we can potentially charge max battery power.
		// For simplicity, we just assign whole hours to "charge" until we have enough slots
		// Assuming max charge power is roughly capacity / 2 (0.5C), so 1 hour covers half capacity.
		// Let's just say 1 hour buys 3000Wh as an arbitrary assumption if we don't know the inverter max power.
		// Wait, a better approach: if we need energyToBuyWh, each hour can provide at most (Capacity / 2) Wh roughly, or we just keep assigning cheapest hours until we hit total capacity.
		// Let's simplify: assign the N cheapest hours until we hit the energy limit assuming a 3kW average charge rate.
		assumedChargeRateW := 3000.0
		energyToBuyWh -= assumedChargeRateW
	}

	// 6. Build Forecast Map
	for _, hd := range hourlyData {
		ts := hd.Timestamp.Unix()
		if chargeSlots[ts] {
			forecast.HourlyActions[ts] = "charge"
		} else if expensiveSlots[ts] && hd.NetLoadW > 0 {
			// Ensure arbitrage spread is profitable (e.g. > 20% loss compensation)
			// Let's check the cheapest slot we bought at to ensure discharge makes sense
			minPrice := sortedByPriceAsc[0].Price
			if hd.Price > minPrice/0.80 {
				forecast.HourlyActions[ts] = "discharge"
			} else {
				forecast.HourlyActions[ts] = "hold"
			}
		} else {
			forecast.HourlyActions[ts] = "hold"
		}
	}

	log.Printf("[INFO] DynamicBatteryForecast: Calculated new plan for next 24h. Target energy: %.1f Wh", targetEnergyWh)
	return forecast
}

func fetchSolarForecast(lat, lon float64, start, end time.Time) []SolarForecastPoint {
	var forecast []SolarForecastPoint
	if lat == 0 && lon == 0 {
		return forecast
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=direct_radiation,diffuse_radiation&forecast_days=2", lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		return forecast
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return forecast
	}

	var meteoData struct {
		Hourly struct {
			Time             []string  `json:"time"`
			DirectRadiation  []float64 `json:"direct_radiation"`
			DiffuseRadiation []float64 `json:"diffuse_radiation"`
		} `json:"hourly"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meteoData); err != nil {
		return forecast
	}

	for i, tStr := range meteoData.Hourly.Time {
		t, err := time.Parse("2006-01-02T15:04", tStr)
		if err != nil {
			continue
		}

		tUTC := t.UTC()
		if tUTC.Before(start) || tUTC.After(end) {
			continue
		}

		direct := 0.0
		diffuse := 0.0
		if i < len(meteoData.Hourly.DirectRadiation) {
			direct = meteoData.Hourly.DirectRadiation[i]
		}
		if i < len(meteoData.Hourly.DiffuseRadiation) {
			diffuse = meteoData.Hourly.DiffuseRadiation[i]
		}

		forecast = append(forecast, SolarForecastPoint{
			Timestamp:       tUTC,
			EstimatedPowerW: direct + diffuse,
		})
	}

	return forecast
}
