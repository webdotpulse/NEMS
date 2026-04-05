package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type PricePoint struct {
	Timestamp   time.Time `json:"timestamp"`
	PricePerKwh float64   `json:"price_per_kwh"`
}

type TariffProvider interface {
	FetchPrices(start, end time.Time) ([]PricePoint, error)
}

type EnergyZeroProvider struct{}

func (p *EnergyZeroProvider) FetchPrices(start, end time.Time) ([]PricePoint, error) {
	url := fmt.Sprintf("https://api.energyzero.nl/v1/energyprices?fromDate=%s&tillDate=%s&interval=4&usageType=1&inclBtw=true",
		start.UTC().Format("2006-01-02T15:04:05.000Z"),
		end.UTC().Format("2006-01-02T15:04:05.000Z"))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("energyzero api returned status: %d", resp.StatusCode)
	}

	var result struct {
		Prices []struct {
			ReadingDate time.Time `json:"readingDate"`
			Price       float64   `json:"price"`
		} `json:"Prices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var points []PricePoint
	for _, rp := range result.Prices {
		points = append(points, PricePoint{
			Timestamp:   rp.ReadingDate,
			PricePerKwh: rp.Price,
		})
	}
	return points, nil
}

type TariffManager struct {
	provider TariffProvider
	stopCh   chan struct{}
}

var TariffMgr *TariffManager

func InitTariffManager() {
	TariffMgr = &TariffManager{
		provider: &EnergyZeroProvider{},
		stopCh:   make(chan struct{}),
	}
}

func (tm *TariffManager) Start() {
	go func() {
		// Try to fetch immediately if we are missing tomorrow's data and it's past 13:30 CET
		tm.fetchIfNeeded()

		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				tm.fetchIfNeeded()
			case <-tm.stopCh:
				return
			}
		}
	}()
}

func (tm *TariffManager) Stop() {
	close(tm.stopCh)
}

func (tm *TariffManager) fetchIfNeeded() {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		log.Printf("[ERROR] TariffManager: error loading location: %v", err)
		loc = time.UTC
	}

	now := time.Now().In(loc)

	// We fetch if it's past 13:30 CET
	// And we only need to fetch once a day, so we check if tomorrow's 23:00 data is in the DB
	// We fetch if it's past 13:30 CET. However, if we don't even have today's data, we should fetch immediately.
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	var countToday int
	_ = db.QueryRow("SELECT count(*) FROM epex_prices WHERE timestamp >= ?", startOfToday.UTC()).Scan(&countToday)

	if countToday == 0 {
		// Fetch immediately if we have zero data for today (e.g. fresh install)
		log.Println("[INFO] TariffManager: No data for today found, bypassing 13:30 check")
	} else if now.Hour() < 13 || (now.Hour() == 13 && now.Minute() < 30) {
		return // Too early, EPEX prices for tomorrow aren't published yet (usually 13:00 or 13:30)
	}

	tomorrowEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, loc).AddDate(0, 0, 1)

	var count int
	err = db.QueryRow("SELECT count(*) FROM epex_prices WHERE timestamp = ?", tomorrowEnd.UTC()).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] TariffManager: error checking epex_prices: %v", err)
		return
	}

	if count > 0 {
		// We already have tomorrow's data
		return
	}

	log.Println("[INFO] TariffManager: Fetching EPEX prices...")

	// Fetch today and tomorrow
	startOfToday = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfTomorrow := startOfToday.AddDate(0, 0, 2).Add(-1 * time.Second)

	prices, err := tm.provider.FetchPrices(startOfToday, endOfTomorrow)
	if err != nil {
		log.Printf("[ERROR] TariffManager: Failed to fetch prices: %v", err)
		go SendWebhookAlert("🚨 Tariff Error: Failed to fetch Day-Ahead energy prices from provider.")
		return
	}

	// Insert into DB
	for _, p := range prices {
		_, err := db.Exec("INSERT OR REPLACE INTO epex_prices (timestamp, price_per_kwh) VALUES (?, ?)", p.Timestamp.UTC(), p.PricePerKwh)
		if err != nil {
			log.Printf("[ERROR] TariffManager: DB insert error: %v", err)
		}
	}

	log.Printf("[INFO] TariffManager: Successfully fetched and stored %d price points", len(prices))
}
