package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"sync"
	"io"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"

	_ "github.com/mattn/go-sqlite3"

	"nems/internal/ocpp"

	"github.com/soheilhy/cmux"
)


// RingBuffer for logs
type LogRingBuffer struct {
	mu     sync.RWMutex
	logs   []string
	maxLen int
	head   int
	count  int
}

func (r *LogRingBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg := string(p)
	r.logs[r.head] = msg
	r.head = (r.head + 1) % r.maxLen
	if r.count < r.maxLen {
		r.count++
	}
	return len(p), nil
}

func (r *LogRingBuffer) GetLogs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]string, 0, r.count)
	if r.count == 0 {
		return res
	}

	start := 0
	if r.count == r.maxLen {
		start = r.head
	}

	for i := 0; i < r.count; i++ {
		res = append(res, r.logs[(start+i)%r.maxLen])
	}
	return res
}

func (r *LogRingBuffer) ClearLogs() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.count = 0
	for i := range r.logs {
		r.logs[i] = ""
	}
}

var logBuffer *LogRingBuffer
var BuildNumber = "development"

func InitLogger() {
	logBuffer = &LogRingBuffer{
		logs:   make([]string, 1000),
		maxLen: 1000,
	}
	multiWriter := io.MultiWriter(os.Stdout, logBuffer)
	log.SetOutput(multiWriter)
}

func ensureCertificates(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); err == nil {
		return nil
	}

	log.Println("[INFO] Generating self-signed certificate...")
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

	var ips []net.IP
	ips = append(ips, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err == nil {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip != nil {
						ips = append(ips, ip)
					}
				}
			}
		}
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
		IPAddresses:           ips,
		DNSNames:              []string{"localhost", "ems", "ems.local"},
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

	log.Println("[INFO] Self-signed certificate generated.")
	return nil
}

var db *sql.DB

func ensureColumnExists(db *sql.DB, tableName, colName, colDef string) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[ERROR] Failed to check columns for table %s: %v", tableName, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err == nil {
			if name == colName {
				return // Column exists
			}
		}
	}

	alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, colName, colDef)
	_, err = db.Exec(alterQuery)
	if err != nil {
		log.Printf("[ERROR] Failed to add column %s to %s: %v", colName, tableName, err)
	} else {
		log.Printf("[INFO] Added missing column %s to table %s", colName, tableName)
	}
}

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

// main is the application entry point. It initializes the logger,
// opens the local SQLite database, sets up WAL mode for concurrency and SD-card wear reduction,
// creates tables, registers all HTTP handlers, starts background polling routines,
// and finally binds the HTTP/HTTPS servers.
func main() {
	InitLogger()

	var err error
	db, err = sql.Open("sqlite3", "./nems.db")
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	defer db.Close()

	// Ensure connection is established
	err = db.Ping()
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
	log.Println("[INFO] Connected to SQLite database")

	// Set connection pool for WAL mode safety
	db.SetMaxOpenConns(1)

	// Configure WAL mode for SD-card optimization
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal("[FATAL] Failed to set WAL mode:", err)
	}
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		log.Fatal("[FATAL] Failed to set synchronous mode:", err)
	}
	_, err = db.Exec("PRAGMA temp_store=MEMORY;")
	if err != nil {
		log.Fatal("[FATAL] Failed to set temp_store:", err)
	}
	log.Println("[INFO] SQLite WAL mode, synchronous=NORMAL, and temp_store=MEMORY enabled for SD card optimization")

	// Create measurements table
	createMeasurementsSQL := `
	CREATE TABLE IF NOT EXISTS measurements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		device_id TEXT NOT NULL,
		power_w REAL NOT NULL,
		battery_power_w REAL NOT NULL DEFAULT 0,
		grid_power_w REAL NOT NULL DEFAULT 0,
		energy_kwh REAL NOT NULL
	);
	`
	_, err = db.Exec(createMeasurementsSQL)
	if err != nil {
		log.Fatal("[FATAL] Failed to create measurements table:", err)
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
		battery_capacity REAL DEFAULT 0,
		inverter_rated_power_kw REAL DEFAULT 0,
		ocpp_proxy_url TEXT DEFAULT ''
	);
	`
	_, err = db.Exec(createDevicesSQL)
	if err != nil {
		log.Fatal("[FATAL] Failed to create devices table:", err)
	}

	// Add new columns to measurements (for existing databases)
	ensureColumnExists(db, "measurements", "battery_power_w", "REAL NOT NULL DEFAULT 0")
	ensureColumnExists(db, "measurements", "grid_power_w", "REAL NOT NULL DEFAULT 0")

	// Add new columns if they don't exist (for existing databases)
	ensureColumnExists(db, "devices", "username", "TEXT DEFAULT ''")
	ensureColumnExists(db, "devices", "password", "TEXT DEFAULT ''")
	ensureColumnExists(db, "devices", "has_grid_meter", "BOOLEAN DEFAULT 0")
	ensureColumnExists(db, "devices", "has_battery", "BOOLEAN DEFAULT 0")
	ensureColumnExists(db, "devices", "battery_capacity", "REAL DEFAULT 0")
	ensureColumnExists(db, "devices", "charge_mode", "TEXT DEFAULT 'eco'")
	ensureColumnExists(db, "devices", "battery_mode", "TEXT DEFAULT 'auto'")
	ensureColumnExists(db, "devices", "inverter_rated_power_kw", "REAL DEFAULT 0")
	ensureColumnExists(db, "devices", "ocpp_proxy_url", "TEXT DEFAULT ''")

	createEpexPricesSQL := `
	CREATE TABLE IF NOT EXISTS epex_prices (
		timestamp DATETIME PRIMARY KEY,
		price_per_kwh REAL NOT NULL
	);
	`
	_, err = db.Exec(createEpexPricesSQL)
	if err != nil {
		log.Fatal("[FATAL] Failed to create epex_prices table:", err)
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
		log.Fatal("[FATAL] Failed to create site_settings table:", err)
	}
	// Insert default if not exists
	_, _ = db.Exec("INSERT OR IGNORE INTO site_settings (id, strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment) VALUES (1, 'eco', 2.5, 0)")

	ensureColumnExists(db, "site_settings", "battery_grid_charge_strategy", "TEXT DEFAULT 'price_only'")
	ensureColumnExists(db, "site_settings", "force_charge_below_euro", "REAL DEFAULT 0")
	ensureColumnExists(db, "site_settings", "force_discharge_above_euro", "REAL DEFAULT 999.0")
	ensureColumnExists(db, "site_settings", "smart_ev_cheapest_hours", "INTEGER DEFAULT 0")
	ensureColumnExists(db, "site_settings", "grid_nominal_current_a", "REAL DEFAULT 25.0")
	ensureColumnExists(db, "site_settings", "grid_system", "TEXT DEFAULT 'single_phase_230v'")
	ensureColumnExists(db, "site_settings", "allowed_grid_import_kw", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "allowed_grid_export_kw", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "appliance_turn_on_excess_w", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "peak_shaving_buffer_w", "REAL DEFAULT 200.0")
	ensureColumnExists(db, "site_settings", "peak_shaving_rampup_w", "REAL DEFAULT 500.0")
	ensureColumnExists(db, "site_settings", "timezone", "TEXT DEFAULT 'Europe/Brussels'")
	ensureColumnExists(db, "site_settings", "latitude", "REAL DEFAULT 50.8503")
	ensureColumnExists(db, "site_settings", "longitude", "REAL DEFAULT 4.3517")

	ensureColumnExists(db, "site_settings", "contract_type", "TEXT DEFAULT 'dynamic'")
	ensureColumnExists(db, "site_settings", "fixed_price_peak_kwh", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "fixed_price_off_peak_kwh", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "fixed_inject_price_kwh", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "dynamic_markup_kwh", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "engie_markup_peak", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "engie_markup_off_peak", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "engie_markup_super_off_peak", "REAL DEFAULT 0.0")
	ensureColumnExists(db, "site_settings", "engie_multiplier", "REAL DEFAULT 1.0")
	ensureColumnExists(db, "site_settings", "engie_base_fee", "REAL DEFAULT 0.0")

	ensureColumnExists(db, "site_settings", "custom_charge_schedule", "TEXT DEFAULT '[]'")
	ensureColumnExists(db, "site_settings", "superdal_optimization_enabled", "BOOLEAN DEFAULT 0")
	ensureColumnExists(db, "site_settings", "superdal_target_soc", "REAL DEFAULT 100.0")

	log.Println("[INFO] Database schema initialized")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", handleStatus)

	mux.HandleFunc("/api/live", handleLiveStream)
	mux.HandleFunc("/api/daily", handleDailyAggregates)
	mux.HandleFunc("/api/energy", handleEnergy)
	mux.HandleFunc("/api/logs", handleLogs)
	mux.HandleFunc("/api/system/info", handleSystemInfo)
	mux.HandleFunc("/api/system/reboot", handleSystemReboot)
	mux.HandleFunc("/api/system/reset-db", handleSystemResetDb)
	mux.HandleFunc("/api/network/scan", handleNetworkScan)

	mux.HandleFunc("/api/tariffs/today", handleTariffsToday)
	mux.HandleFunc("/api/tariffs/forecast", handleTariffForecast)
	mux.HandleFunc("/api/solar/forecast", handleSolarForecast)
	mux.HandleFunc("/api/settings", handleSettings)
	mux.HandleFunc("/api/templates", handleTemplates)
	mux.HandleFunc("/api/devices", handleDevices)
	mux.HandleFunc("/api/history", handleHistory)
	mux.HandleFunc("/api/devices/", handleDevice)

	// OCPP WebSocket endpoint
	ocpp.GetDeviceProxyUrl = func(chargePointId string) string {
		if PollerMgr == nil {
			return ""
		}
		for _, data := range PollerMgr.GetDeviceCache() {
			// Find the device matching this host (ChargePoint ID) and return its Proxy URL
			if data.Category == "charger" {
				for _, p := range PollerMgr.GetPollers() {
					if p.GetDevice().Host == chargePointId {
						return p.GetDevice().OcppProxyUrl
					}
				}
			}
		}
		return ""
	}
	mux.HandleFunc("/api/ocpp/", ocpp.HandleWebSocket)

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


	handler := enableCORS(mux)

	err = ensureCertificates("cert.pem", "key.pem")
	if err != nil {
		log.Printf("[ERROR] Failed to ensure certificates: %v. HTTPS disabled.", err)
	}

	// Helper to start multiplexed server on a given port
	startMultiplexedServer := func(port string, isPrimary bool) {
		l, err := net.Listen("tcp", port)
		if err != nil {
			if isPrimary {
				log.Fatalf("[FATAL] Failed to bind to primary port %s: %v", port, err)
			} else {
				log.Printf("[WARN] Failed to bind to optional port %s: %v", port, err)
				return
			}
		}

		m := cmux.New(l)
		httpsL := m.Match(cmux.TLS())
		httpL := m.Match(cmux.Any())

		// Start HTTP server
		go func() {
			if err := http.Serve(httpL, handler); err != nil && err != http.ErrServerClosed {
				log.Printf("[ERROR] HTTP server on %s failed: %v", port, err)
			}
		}()

		// Start HTTPS server if certificates exist
		if _, errCert := os.Stat("cert.pem"); errCert == nil {
			go func() {
				srv := &http.Server{Handler: handler}
				if err := srv.ServeTLS(httpsL, "cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
					log.Printf("[ERROR] HTTPS server on %s failed: %v", port, err)
				}
			}()
		}

		// Start serving
		go func() {
			log.Printf("[INFO] Starting multiplexed HTTP/HTTPS server on %s", port)
			if err := m.Serve(); err != nil && err != net.ErrClosed {
				log.Printf("[ERROR] Multiplexer on %s failed: %v", port, err)
			}
		}()
	}

	// Always start on 8080 (Primary port)
	startMultiplexedServer(":8080", true)

	// Start on 80 and 443 (Optional, may require root)
	startMultiplexedServer(":80", false)
	startMultiplexedServer(":443", false)

	select {} // Block main thread

}
