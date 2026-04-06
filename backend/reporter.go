package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"image/color"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type Reporter struct {
	stopCh chan struct{}
}

var GlobalReporter *Reporter

func InitReporter() {
	GlobalReporter = &Reporter{
		stopCh: make(chan struct{}),
	}

	go func() {
		// Run every hour to check if it's the right time
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		var lastRunWeek int

		for {
			select {
			case <-ticker.C:
				loc, err := time.LoadLocation("Europe/Amsterdam")
				if err != nil {
					loc = time.UTC
				}
				now := time.Now().In(loc)

				// Run on Sunday between 23:00 and 23:59
				_, currentWeek := now.ISOWeek()
				if now.Weekday() == time.Sunday && now.Hour() == 23 && lastRunWeek != currentWeek {
					// Check if reporting is enabled
					var enabled bool
					err := db.QueryRow("SELECT weekly_report_enabled FROM site_settings WHERE id = 1").Scan(&enabled)
					if err == nil && enabled {
						log.Println("[INFO] Reporter: Generating and sending weekly report...")
						GenerateAndSendWeeklyReport()
					}
					lastRunWeek = currentWeek
				}
			case <-GlobalReporter.stopCh:
				log.Println("[INFO] Reporter: Background task stopped")
				return
			}
		}
	}()
}

func (r *Reporter) Stop() {
	close(r.stopCh)
}

// GeneratePDFReport creates a PDF report for a given period: "daily", "weekly", "monthly", "yearly".
func GeneratePDFReport(period string) ([]byte, error) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	now := time.Now().In(loc)
	var startDate, endDate time.Time

	switch period {
	case "daily":
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		startDate = endDate.AddDate(0, 0, -1)
	case "weekly":
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		startDate = endDate.AddDate(0, 0, -7)
	case "monthly":
		endDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).UTC()
		startDate = endDate.AddDate(0, -1, 0)
	case "yearly":
		endDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc).UTC()
		startDate = endDate.AddDate(-1, 0, 0)
	case "all":
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, loc).UTC()
	default: // fallback to weekly
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		startDate = endDate.AddDate(0, 0, -7)
	}

	query := `
		WITH minute_aggs AS (
			SELECT
				strftime('%Y-%m-%d', m.timestamp) as bucket,
				SUM(CASE WHEN (d.template IN ('huawei_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (d.template IN ('huawei_inverter', 'enerlution_inverter') AND d.has_grid_meter = 1)) AND m.grid_power_w > 0 THEN (m.grid_power_w / 60000.0) ELSE 0 END) as grid_import,
				SUM(CASE WHEN (d.template IN ('huawei_dongle', 'homewizard_meter', 'p1_serial', 'p1_network') OR (d.template IN ('huawei_inverter', 'enerlution_inverter') AND d.has_grid_meter = 1)) AND m.grid_power_w < 0 THEN ABS(m.grid_power_w / 60000.0) ELSE 0 END) as grid_export,
				SUM(CASE WHEN d.template IN ('huawei_inverter', 'solis_inverter', 'sma_inverter', 'enerlution_inverter') AND m.power_w > 0 THEN (m.power_w / 60000.0) ELSE 0 END) as solar_yield,
				SUM(CASE WHEN ((d.template IN ('huawei_inverter', 'enerlution_inverter') AND d.has_battery = 1)) AND m.battery_power_w < 0 THEN ABS(m.battery_power_w / 60000.0) ELSE 0 END) as battery_charge,
				SUM(CASE WHEN ((d.template IN ('huawei_inverter', 'enerlution_inverter') AND d.has_battery = 1)) AND m.battery_power_w > 0 THEN (m.battery_power_w / 60000.0) ELSE 0 END) as battery_discharge
			FROM measurements m
			JOIN devices d ON CAST(m.device_id AS INTEGER) = d.id
			WHERE m.timestamp >= ? AND m.timestamp < ?
			GROUP BY strftime('%Y-%m-%dT%H:%M:00Z', m.timestamp), bucket
		),
		bucket_aggs AS (
			SELECT
				bucket,
				SUM(grid_import) as grid_import,
				SUM(grid_export) as grid_export,
				SUM(solar_yield) as solar_yield,
				SUM(battery_charge) as battery_charge,
				SUM(battery_discharge) as battery_discharge
			FROM minute_aggs
			GROUP BY bucket
		)
		SELECT
			bucket,
			grid_import,
			grid_export,
			solar_yield,
			battery_charge,
			battery_discharge
		FROM bucket_aggs
		ORDER BY bucket ASC
	`

	rows, err := db.Query(query, startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("[ERROR] Reporter: Query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var totalGridImport, totalGridExport, totalSolarYield, totalConsumption float64
	var days []string
	var solarVals plotter.Values
	var importVals plotter.Values
	var consVals plotter.Values

	for rows.Next() {
		var bucket string
		var gImport, gExport, sYield, bCharge, bDischarge sql.NullFloat64

		if err := rows.Scan(&bucket, &gImport, &gExport, &sYield, &bCharge, &bDischarge); err != nil {
			log.Printf("[ERROR] Reporter: Scan failed: %v", err)
			continue
		}

		days = append(days, bucket)

		importVal := 0.0
		exportVal := 0.0
		solarVal := 0.0
		chargeVal := 0.0
		dischargeVal := 0.0

		if gImport.Valid {
			importVal = gImport.Float64
			totalGridImport += importVal
		}
		if gExport.Valid {
			exportVal = gExport.Float64
			totalGridExport += exportVal
		}
		if sYield.Valid {
			solarVal = sYield.Float64
			totalSolarYield += solarVal
		}
		if bCharge.Valid {
			chargeVal = bCharge.Float64
		}
		if bDischarge.Valid {
			dischargeVal = bDischarge.Float64
		}

		solarVals = append(solarVals, solarVal)
		importVals = append(importVals, importVal)

		cons := importVal + solarVal + dischargeVal - exportVal - chargeVal
		if cons < 0 {
			cons = 0
		}
		consVals = append(consVals, cons)
		totalConsumption += cons
	}

	selfSufficiency := 0.0
	if totalConsumption > 0 {
		fromPV := totalConsumption - totalGridImport
		if fromPV < 0 {
			fromPV = 0
		}
		selfSufficiency = (fromPV / totalConsumption) * 100.0
	}

	// Generate Chart
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s Energy Overview", strings.Title(period))
	p.X.Label.Text = "Date"
	p.Y.Label.Text = "Energy (kWh)"

	if len(days) > 0 {
		p.NominalX(days...)
	}

	w := vg.Points(15)

	if len(solarVals) == 0 {
		solarVals = append(solarVals, 0)
		importVals = append(importVals, 0)
		consVals = append(consVals, 0)
	}

	barsSolar, _ := plotter.NewBarChart(solarVals, w)
	barsSolar.Color = color.RGBA{R: 255, G: 200, B: 0, A: 255}
	barsSolar.Offset = -w

	barsImport, _ := plotter.NewBarChart(importVals, w)
	barsImport.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	barsImport.Offset = 0

	barsCons, _ := plotter.NewBarChart(consVals, w)
	barsCons.Color = color.RGBA{R: 255, G: 100, B: 0, A: 255}
	barsCons.Offset = w

	p.Add(barsSolar, barsImport, barsCons)
	p.Legend.Add("Solar Yield", barsSolar)
	p.Legend.Add("Grid Import", barsImport)
	p.Legend.Add("Consumption", barsCons)

	var imgBuf bytes.Buffer
	wt, err := p.WriterTo(6*vg.Inch, 4*vg.Inch, "png")
	if err != nil {
		log.Printf("[ERROR] Reporter: Plot error: %v", err)
		return nil, err
	}
	wt.WriteTo(&imgBuf)

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Pulse EMS - %s Energy Report", strings.Title(period)))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("Report Period: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.CellFormat(0, 8, fmt.Sprintf("Total Solar Generation: %.2f kWh", totalSolarYield), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Total Grid Import: %.2f kWh", totalGridImport), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Total Grid Export: %.2f kWh", totalGridExport), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Total Consumption: %.2f kWh", totalConsumption), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Self-Sufficiency Ratio: %.1f%%", selfSufficiency), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	pdf.RegisterImageOptionsReader("chart", gofpdf.ImageOptions{ImageType: "PNG"}, &imgBuf)
	pdf.ImageOptions("chart", 15, 100, 180, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	var pdfBuf bytes.Buffer
	err = pdf.Output(&pdfBuf)
	if err != nil {
		log.Printf("[ERROR] Reporter: PDF Output error: %v", err)
		return nil, err
	}

	return pdfBuf.Bytes(), nil
}

func GenerateAndSendWeeklyReport() {
	pdfBytes, err := GeneratePDFReport("weekly")
	if err != nil {
		log.Printf("[ERROR] Reporter: Could not generate weekly report PDF: %v", err)
		return
	}

	// Fetch SMTP Settings
	var email, host, username, password, sender string
	var port int
	err = db.QueryRow("SELECT report_email, smtp_host, smtp_port, smtp_username, smtp_password, smtp_sender FROM site_settings WHERE id = 1").
		Scan(&email, &host, &port, &username, &password, &sender)
	if err != nil || host == "" || email == "" {
		log.Printf("[ERROR] Reporter: Missing SMTP settings or email.")
		return
	}

	// Send Email
	auth := smtp.PlainAuth("", username, password, host)
	to := []string{email}

	boundary := "nems-boundary"
	headers := "From: " + sender + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: Pulse EMS - Weekly Energy Report\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n\r\n"

	body := "--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n" +
		"Please find attached your weekly energy report from Pulse EMS.\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: application/pdf; name=\"weekly-report.pdf\"\r\n" +
		"Content-Disposition: attachment; filename=\"weekly-report.pdf\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n"

	encoder := bytes.NewBuffer(pdfBytes)

	// Implement base64 encoding (simplest way is encoding/base64, let's use standard package)

	importStr := make([]byte, 0)
	importStr = append(importStr, []byte(headers)...)
	importStr = append(importStr, []byte(body)...)

	// Base64 encode using stdlib
	b64Encoded := base64.StdEncoding.EncodeToString(encoder.Bytes())

	// Split into 76 char lines
	for i := 0; i < len(b64Encoded); i += 76 {
		end := i + 76
		if end > len(b64Encoded) {
			end = len(b64Encoded)
		}
		importStr = append(importStr, []byte(b64Encoded[i:end])...)
		importStr = append(importStr, []byte("\r\n")...)
	}

	importStr = append(importStr, []byte("--"+boundary+"--\r\n")...)

	// Connect to SMTP with TLS if port is 465
	addr := fmt.Sprintf("%s:%d", host, port)
	if port == 465 {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         host,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			log.Printf("[ERROR] Reporter: Failed to dial SMTP over TLS: %v", err)
			return
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			log.Printf("[ERROR] Reporter: SMTP Client error: %v", err)
			return
		}
		if err = client.Auth(auth); err != nil {
			log.Printf("[ERROR] Reporter: SMTP Auth error: %v", err)
			return
		}
		if err = client.Mail(sender); err != nil {
			log.Printf("[ERROR] Reporter: SMTP Mail error: %v", err)
			return
		}
		if err = client.Rcpt(email); err != nil {
			log.Printf("[ERROR] Reporter: SMTP Rcpt error: %v", err)
			return
		}
		w, err := client.Data()
		if err != nil {
			log.Printf("[ERROR] Reporter: SMTP Data error: %v", err)
			return
		}
		_, err = w.Write(importStr)
		if err != nil {
			log.Printf("[ERROR] Reporter: SMTP Write error: %v", err)
			return
		}
		err = w.Close()
		if err != nil {
			log.Printf("[ERROR] Reporter: SMTP Close error: %v", err)
			return
		}
		client.Quit()
	} else {
		err = smtp.SendMail(addr, auth, sender, to, importStr)
		if err != nil {
			log.Printf("[ERROR] Reporter: SMTP SendMail error: %v", err)
			return
		}
	}

	log.Println("[INFO] Reporter: Weekly report sent successfully.")
}
