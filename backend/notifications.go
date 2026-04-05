package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// SendWebhookAlert reads the webhook URL from the site_settings table and sends a generic text payload.
func SendWebhookAlert(message string) {
	var webhookURL string
	err := db.QueryRow("SELECT alert_webhook_url FROM site_settings WHERE id = 1").Scan(&webhookURL)
	if err != nil || webhookURL == "" {
		return // No webhook configured or error fetching
	}

	payload := map[string]string{
		"text": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[ERROR] Webhook: Failed to marshal payload: %v", err)
		return
	}

	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[ERROR] Webhook: Failed to send alert: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("[ERROR] Webhook: Received non-2xx status code: %d", resp.StatusCode)
		}
	}()
}
