package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"os"
	"net"
	"sync"
	"io"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"

	"nems/internal/models"
	"nems/internal/templates"

	_ "github.com/mattn/go-sqlite3"
)


// RingBuffer for logs
type LogRingBuffer struct {
	mu     sync.RWMutex
	logs   []string
	maxLen int
}

func (r *LogRingBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg := string(p)
	if len(r.logs) >= r.maxLen {
		r.logs = r.logs[1:]
	}
	r.logs = append(r.logs, msg)
	return len(p), nil
}

func (r *LogRingBuffer) GetLogs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]string, len(r.logs))
	copy(res, r.logs)
	return res
}

var logBuffer *LogRingBuffer

func InitLogger() {
	logBuffer = &LogRingBuffer{
		logs:   make([]string, 0, 1000),
		maxLen: 1000,
	}
	multiWriter := io.MultiWriter(os.Stdout, logBuffer)
	log.SetOutput(multiWriter)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logs := logBuffer.GetLogs()
	json.NewEncoder(w).Encode(logs)
}

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	var primaryIP string
	var primaryNetmask string

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					primaryIP = ipnet.IP.String()
					mask := ipnet.Mask
					primaryNetmask = net.IP(mask).String()
					break
				}
			}
		}
	}

	info := map[string]string{
		"hostname": hostname,
		"ip":       primaryIP,
		"netmask":  primaryNetmask,
	}

	json.NewEncoder(w).Encode(info)
}

func ensureCertificates(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		return nil
	}

	log.Println("Generating self-signed certificate...")
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Pulse EMS"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	log.Println("Self-signed certificate generated.")
	return nil
}

var db *sql.DB

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	InitLogger()

	var err error
	db, err = sql.Open("sqlite3", "./nems.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ensure connection is established
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to SQLite database")

	// Configure WAL mode for SD-card optimization
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal("Failed to set WAL mode:", err)
	}
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		log.Fatal("Failed to set synchronous mode:", err)
	}
	_, err = db.Exec("PRAGMA temp_store=MEMORY;")
	if err != nil {
		log.Fatal("Failed to set temp_store:", err)
	}
	log.Println("SQLite WAL mode, synchronous=NORMAL, and temp_store=MEMORY enabled for SD card optimization")

	// Create measurements table
	createMeasurementsSQL := `
	CREATE TABLE IF NOT EXISTS measurements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		device_id TEXT NOT NULL,
		power_w REAL NOT NULL,
		energy_kwh REAL NOT NULL
	);
	`
	_, err = db.Exec(createMeasurementsSQL)
	if err != nil {
		log.Fatal("Failed to create measurements table:", err)
	}

	createDevicesSQL := `
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		template TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		modbus_id INTEGER NOT NULL,
		username TEXT DEFAULT '',
		password TEXT DEFAULT '',
		has_grid_meter BOOLEAN DEFAULT 0,
		has_battery BOOLEAN DEFAULT 0,
		battery_capacity REAL DEFAULT 0
	);
	`
	_, err = db.Exec(createDevicesSQL)
	if err != nil {
		log.Fatal("Failed to create devices table:", err)
	}

	// Add new columns if they don't exist (for existing databases)
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN username TEXT DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN password TEXT DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN has_grid_meter BOOLEAN DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN has_battery BOOLEAN DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN battery_capacity REAL DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN charge_mode TEXT DEFAULT 'eco'")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN battery_mode TEXT DEFAULT 'auto'")

	createEpexPricesSQL := `
	CREATE TABLE IF NOT EXISTS epex_prices (
		timestamp DATETIME PRIMARY KEY,
		price_per_kwh REAL NOT NULL
	);
	`
	_, err = db.Exec(createEpexPricesSQL)
	if err != nil {
		log.Fatal("Failed to create epex_prices table:", err)
	}

	createSettingsSQL := `
	CREATE TABLE IF NOT EXISTS site_settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		strategy_mode TEXT DEFAULT 'eco',
		capacity_peak_limit_kw REAL DEFAULT 2.5,
		active_inverter_curtailment BOOLEAN DEFAULT 0
	);
	`
	_, err = db.Exec(createSettingsSQL)
	if err != nil {
		log.Fatal("Failed to create site_settings table:", err)
	}
	// Insert default if not exists
	_, _ = db.Exec("INSERT OR IGNORE INTO site_settings (id, strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment) VALUES (1, 'eco', 2.5, 0)")

	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN force_charge_below_euro REAL DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN smart_ev_cheapest_hours INTEGER DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN grid_nominal_current_a REAL DEFAULT 25.0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN grid_system TEXT DEFAULT 'single_phase_230v'")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN allowed_grid_import_kw REAL DEFAULT 0.0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN allowed_grid_export_kw REAL DEFAULT 0.0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN appliance_turn_on_excess_w REAL DEFAULT 0.0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN peak_shaving_buffer_w REAL DEFAULT 200.0")
	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN peak_shaving_rampup_w REAL DEFAULT 500.0")

	log.Println("Database schema initialized")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("/api/live", handleLiveStream)
	mux.HandleFunc("/api/daily", handleDailyAggregates)
	mux.HandleFunc("/api/logs", handleLogs)
	mux.HandleFunc("/api/system/info", handleSystemInfo)
	mux.HandleFunc("/api/network/scan", handleNetworkScan)

	mux.HandleFunc("/api/tariffs/today", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		loc, err := time.LoadLocation("Europe/Amsterdam")
		if err != nil {
			loc = time.UTC
		}
		now := time.Now().In(loc)
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		endOfTomorrow := time.Date(now.Year(), now.Month(), now.Day()+2, 0, 0, 0, 0, loc).UTC()

		rows, err := db.Query("SELECT timestamp, price_per_kwh FROM epex_prices WHERE timestamp >= ? AND timestamp < ? ORDER BY timestamp ASC", startOfToday, endOfTomorrow)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var prices []PricePoint
		for rows.Next() {
			var p PricePoint
			var ts time.Time
			if err := rows.Scan(&ts, &p.PricePerKwh); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			p.Timestamp = ts.In(loc)
			prices = append(prices, p)
		}

		if prices == nil {
			prices = []PricePoint{}
		}
		json.NewEncoder(w).Encode(prices)
	})

	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, force_charge_below_euro, smart_ev_cheapest_hours, grid_nominal_current_a, grid_system, allowed_grid_import_kw, allowed_grid_export_kw, appliance_turn_on_excess_w, peak_shaving_buffer_w, peak_shaving_rampup_w FROM site_settings WHERE id = 1")
			var settings models.SiteSettings
			err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.ForceChargeBelowEuro, &settings.SmartEvCheapestHours, &settings.GridNominalCurrentA, &settings.GridSystem, &settings.AllowedGridImportKw, &settings.AllowedGridExportKw, &settings.ApplianceTurnOnExcessW, &settings.PeakShavingBufferW, &settings.PeakShavingRampupW)
			if err != nil {
				if err == sql.ErrNoRows {
					// Fallback
					settings = models.SiteSettings{
						StrategyMode:              "eco",
						CapacityPeakLimitKw:       2.5,
						ActiveInverterCurtailment: false,
						ForceChargeBelowEuro:      0.0,
						SmartEvCheapestHours:      0,
						GridNominalCurrentA:       25.0,
						GridSystem:                "single_phase_230v",
						AllowedGridImportKw:       0.0,
						AllowedGridExportKw:       0.0,
						ApplianceTurnOnExcessW:    0.0,
						PeakShavingBufferW:        200.0,
						PeakShavingRampupW:        500.0,
					}
					json.NewEncoder(w).Encode(settings)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(settings)
		} else if r.Method == "PUT" {
			var settings models.SiteSettings
			if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			_, err = db.Exec("UPDATE site_settings SET strategy_mode = ?, capacity_peak_limit_kw = ?, active_inverter_curtailment = ?, force_charge_below_euro = ?, smart_ev_cheapest_hours = ?, grid_nominal_current_a = ?, grid_system = ?, allowed_grid_import_kw = ?, allowed_grid_export_kw = ?, appliance_turn_on_excess_w = ?, peak_shaving_buffer_w = ?, peak_shaving_rampup_w = ? WHERE id = 1",
				settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment, settings.ForceChargeBelowEuro, settings.SmartEvCheapestHours, settings.GridNominalCurrentA, settings.GridSystem, settings.AllowedGridImportKw, settings.AllowedGridExportKw, settings.ApplianceTurnOnExcessW, settings.PeakShavingBufferW, settings.PeakShavingRampupW)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(settings)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tList := templates.GetTemplates()
		json.NewEncoder(w).Encode(tList)
	})

	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, charge_mode, battery_mode FROM devices")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var devices []models.Device
			for rows.Next() {
				var d models.Device
				var username sql.NullString
				var password sql.NullString
				var chargeMode sql.NullString
				var batteryMode sql.NullString
				if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password, &d.HasGridMeter, &d.HasBattery, &d.BatteryCapacity, &chargeMode, &batteryMode); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if username.Valid {
					d.Username = username.String
				}
				if password.Valid {
					d.Password = password.String
				}
				if chargeMode.Valid {
					d.ChargeMode = chargeMode.String
				}
				if batteryMode.Valid {
					d.BatteryMode = batteryMode.String
				}

				// Set dynamic status if poller exists
				d.Status = "offline"
				if PollerMgr != nil {
					PollerMgr.mu.Lock()
					if poller, ok := PollerMgr.pollers[d.ID]; ok {
						d.Status = poller.Status()
					}
					PollerMgr.mu.Unlock()
				}

				devices = append(devices, d)
			}
			// ensure non-nil slice in json
			if devices == nil {
				devices = []models.Device{}
			}
			json.NewEncoder(w).Encode(devices)
		} else if r.Method == "POST" {
			var d models.Device
			if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if d.ChargeMode == "" {
				d.ChargeMode = "eco"
			}
			if d.BatteryMode == "" {
				d.BatteryMode = "auto"
			}

			result, err := db.Exec("INSERT INTO devices (name, template, host, port, modbus_id, username, password, has_grid_meter, has_battery, battery_capacity, charge_mode, battery_mode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.ChargeMode, d.BatteryMode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			id, _ := result.LastInsertId()
			d.ID = int(id)

			// Notify poller manager
			if PollerMgr != nil {
				PollerMgr.SyncDevices()
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(d)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/history", handleHistory)

	mux.HandleFunc("/api/devices/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		idStr := strings.TrimPrefix(r.URL.Path, "/api/devices/")
		if strings.HasSuffix(idStr, "/mode") {
			idStr = strings.TrimSuffix(idStr, "/mode")
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid device ID", http.StatusBadRequest)
			return
		}

		if r.Method == "DELETE" {
			_, err = db.Exec("DELETE FROM devices WHERE id = ?", id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Notify poller manager
			if PollerMgr != nil {
				PollerMgr.SyncDevices()
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "deleted"}`))
		} else if strings.HasSuffix(r.URL.Path, "/mode") && r.Method == "PUT" {
			var payload struct {
				ChargeMode  string `json:"charge_mode"`
				BatteryMode string `json:"battery_mode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if payload.ChargeMode != "" {
				_, err = db.Exec("UPDATE devices SET charge_mode = ? WHERE id = ?", payload.ChargeMode, id)
			} else if payload.BatteryMode != "" {
				_, err = db.Exec("UPDATE devices SET battery_mode = ? WHERE id = ?", payload.BatteryMode, id)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "updated"}`))

		} else if r.Method == "PUT" {
			var d models.Device
			if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if d.ChargeMode == "" {
				d.ChargeMode = "eco"
			}
			if d.BatteryMode == "" {
				d.BatteryMode = "auto"
			}

			_, err = db.Exec("UPDATE devices SET name = ?, template = ?, host = ?, port = ?, modbus_id = ?, username = ?, password = ?, has_grid_meter = ?, has_battery = ?, battery_capacity = ?, charge_mode = ?, battery_mode = ? WHERE id = ?",
				d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, d.HasGridMeter, d.HasBattery, d.BatteryCapacity, d.ChargeMode, d.BatteryMode, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			d.ID = id

			// Notify poller manager
			if PollerMgr != nil {
				PollerMgr.SyncDevices()
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(d)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve frontend SPA
	fs := http.FileServer(http.Dir("./dist"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists in the static directory
		if _, err := http.Dir("./dist").Open(r.URL.Path); err != nil {
			// Fallback to index.html for SPA routing if file not found
			http.ServeFile(w, r, "./dist/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Initialize PollerManager
	InitPollerManager()
	if PollerMgr != nil {
		PollerMgr.SyncDevices()
		PollerMgr.Start()
	}

	// Initialize StrategyController
	InitStrategyController()
	if StrategyCtrl != nil {
		StrategyCtrl.Start()
	}

	// Initialize TariffManager
	InitTariffManager()
	if TariffMgr != nil {
		TariffMgr.Start()
	}


	// Try port 80 (requires root/setcap) or fallback to 8080
	httpPort := ":80"
	httpsPort := ":443"

	handler := enableCORS(mux)

	go func() {
		log.Println("Starting HTTP server on :80 (fallback to :8080)")
		err := http.ListenAndServe(httpPort, handler)
		if err != nil {
			log.Printf("Failed to bind to :80 (%v). Trying :8080 instead.", err)
			err = http.ListenAndServe(":8080", handler)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	err = ensureCertificates("cert.pem", "key.pem")
	if err != nil {
		log.Printf("Failed to ensure certificates: %v. HTTPS disabled.", err)
	} else {
		log.Println("Starting HTTPS server on :443 (fallback to :8443)")
		err = http.ListenAndServeTLS(httpsPort, "cert.pem", "key.pem", handler)
		if err != nil {
			log.Printf("Failed to bind to :443 (%v). Trying :8443 instead.", err)
			err = http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", handler)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	select {} // Block main thread

}
