package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// FetchEliaSystemLoad demonstrates integration with the Elia Open Data API
// for features other than imbalance (e.g., retrieving national grid load).
func FetchEliaSystemLoad() {
	url := "https://opendata.elia.be/api/explore/v2.1/catalog/datasets/ods032/records?limit=1"

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ERROR] EliaAPI: Failed to fetch system load: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] EliaAPI: Returned status %d", resp.StatusCode)
		return
	}

	var result struct {
		Results []struct {
			Datetime time.Time `json:"datetime"`
			EliaLoad float64   `json:"elia_load"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[ERROR] EliaAPI: Failed to decode JSON: %v", err)
		return
	}

	if len(result.Results) > 0 {
		log.Printf("[INFO] EliaAPI: Current National System Load at %s is %.1f MW",
			result.Results[0].Datetime.Format(time.RFC3339), result.Results[0].EliaLoad)
	}
}
