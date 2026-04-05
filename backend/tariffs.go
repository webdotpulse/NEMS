package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type PricePoint struct {
	Timestamp   time.Time `json:"timestamp"`
	PricePerKwh float64   `json:"price_per_kwh"`
}

type TariffProvider interface {
	FetchPrices(start, end time.Time) ([]PricePoint, error)
}

type EntsoeProvider struct {
	ApiKey   string
	AreaCode string
}

func (p *EntsoeProvider) FetchPrices(start, end time.Time) ([]PricePoint, error) {
	// Format dates for ENTSO-E (YYYYMMDDHH00) in UTC
	fmtStr := "200601021500"
	startStr := start.UTC().Format(fmtStr)
	endStr := end.UTC().Format(fmtStr)

	// We use A44 documentType for Day Ahead Prices
	url := fmt.Sprintf("https://web-api.tp.entsoe.eu/api?securityToken=%s&documentType=A44&in_Domain=%s&out_Domain=%s&periodStart=%s&periodEnd=%s",
		p.ApiKey, p.AreaCode, p.AreaCode, startStr, endStr)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("entsoe api returned status: %d", resp.StatusCode)
	}

	type Point struct {
		Position int     `xml:"position"`
		Price    float64 `xml:"price.amount"`
	}

	type TimeInterval struct {
		Start string `xml:"start"`
		End   string `xml:"end"`
	}

	type Period struct {
		TimeInterval TimeInterval `xml:"timeInterval"`
		Resolution   string       `xml:"resolution"`
		Points       []Point      `xml:"Point"`
	}

	type TimeSeries struct {
		Period []Period `xml:"Period"`
	}

	type PublicationMarketDocument struct {
		XMLName    xml.Name     `xml:"Publication_MarketDocument"`
		TimeSeries []TimeSeries `xml:"TimeSeries"`
	}

	var doc PublicationMarketDocument
	if err := xml.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	var points []PricePoint
	for _, ts := range doc.TimeSeries {
		for _, p := range ts.Period {
			// ENTSO-E times are in UTC with format 2006-01-02T15:04Z
			periodStart, err := time.Parse("2006-01-02T15:04Z", p.TimeInterval.Start)
			if err != nil {
				continue
			}

			// We assume the resolution is PT60M (1 hour). For day-ahead prices it usually is.
			for _, point := range p.Points {
				// position is 1-indexed
				offset := time.Duration(point.Position-1) * time.Hour
				timestamp := periodStart.Add(offset)

				// Filter to only include the requested range to avoid returning excess data
				if (timestamp.Equal(start) || timestamp.After(start)) && timestamp.Before(end) {
					// Price per MWh -> Price per kWh
					priceKwh := point.Price / 1000.0
					points = append(points, PricePoint{
						Timestamp:   timestamp,
						PricePerKwh: priceKwh,
					})
				}
			}
		}
	}

	return points, nil
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

var currentImbalancePrice float64
var lastImbalanceUpdate time.Time
var imbalancePriceMu sync.RWMutex

func GetCurrentImbalancePrice() float64 {
	imbalancePriceMu.RLock()
	defer imbalancePriceMu.RUnlock()
	return currentImbalancePrice
}

func SetCurrentImbalancePrice(price float64) {
	imbalancePriceMu.Lock()
	defer imbalancePriceMu.Unlock()
	currentImbalancePrice = price
	lastImbalanceUpdate = time.Now()
}

type TariffManager struct {
	provider TariffProvider
	stopCh   chan struct{}
}

var TariffMgr *TariffManager

func InitTariffManager() {
	TariffMgr = &TariffManager{
		provider: nil, // We'll set this dynamically
		stopCh:   make(chan struct{}),
	}
}

func (tm *TariffManager) getProvider() TariffProvider {
	var entsoeKey, entsoeArea string
	err := db.QueryRow("SELECT entsoe_api_key, entsoe_area_code FROM site_settings WHERE id = 1").Scan(&entsoeKey, &entsoeArea)

	if err == nil && entsoeKey != "" {
		if entsoeArea == "" {
			entsoeArea = "10YBE----------2"
		}
		return &EntsoeProvider{
			ApiKey:   entsoeKey,
			AreaCode: entsoeArea,
		}
	}

	return &EnergyZeroProvider{}
}

func (tm *TariffManager) Start() {
	go func() {
		// Try to fetch immediately if we are missing tomorrow's data and it's past 13:30 CET
		tm.fetchIfNeeded()

		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		imbalanceTicker := time.NewTicker(1 * time.Minute)
		defer imbalanceTicker.Stop()

		for {
			select {
			case <-ticker.C:
				tm.fetchIfNeeded()
			case <-imbalanceTicker.C:
				tm.fetchImbalancePrice()
			case <-tm.stopCh:
				return
			}
		}
	}()
}

func (tm *TariffManager) fetchImbalancePrice() {
	resp, err := http.Get("https://opendata.elia.be/api/explore/v2.1/catalog/datasets/ods161/records?order_by=datetime%20DESC&limit=1")
	if err != nil {
		log.Printf("[ERROR] TariffManager: Failed to fetch imbalance price: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] TariffManager: Imbalance API returned status: %d", resp.StatusCode)
		return
	}

	var result struct {
		Results []struct {
			ImbalancePrice float64 `json:"imbalanceprice"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[ERROR] TariffManager: Failed to decode imbalance response: %v", err)
		return
	}

	if len(result.Results) > 0 {
		SetCurrentImbalancePrice(result.Results[0].ImbalancePrice)
	}
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

	provider := tm.getProvider()

	// Fetch today and tomorrow
	startOfToday = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfTomorrow := startOfToday.AddDate(0, 0, 2).Add(-1 * time.Second)

	prices, err := provider.FetchPrices(startOfToday, endOfTomorrow)
	if err != nil {
		log.Printf("[ERROR] TariffManager: Failed to fetch prices: %v", err)
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
