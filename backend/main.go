package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"nems/internal/models"
	"nems/internal/templates"

	_ "github.com/mattn/go-sqlite3"
)

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
		password TEXT DEFAULT ''
	);
	`
	_, err = db.Exec(createDevicesSQL)
	if err != nil {
		log.Fatal("Failed to create devices table:", err)
	}

	// Add new columns if they don't exist (for existing databases)
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN username TEXT DEFAULT ''")
	_, _ = db.Exec("ALTER TABLE devices ADD COLUMN password TEXT DEFAULT ''")

	log.Println("Database schema initialized")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("/api/live", handleLiveStream)

	mux.HandleFunc("/api/templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tList := templates.GetTemplates()
		json.NewEncoder(w).Encode(tList)
	})

	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			rows, err := db.Query("SELECT id, name, template, host, port, modbus_id, username, password FROM devices")
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
				if err := rows.Scan(&d.ID, &d.Name, &d.Template, &d.Host, &d.Port, &d.ModbusID, &username, &password); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if username.Valid {
					d.Username = username.String
				}
				if password.Valid {
					d.Password = password.String
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

			result, err := db.Exec("INSERT INTO devices (name, template, host, port, modbus_id, username, password) VALUES (?, ?, ?, ?, ?, ?, ?)", d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password)
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
		} else if r.Method == "PUT" {
			var d models.Device
			if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			_, err = db.Exec("UPDATE devices SET name = ?, template = ?, host = ?, port = ?, modbus_id = ?, username = ?, password = ? WHERE id = ?",
				d.Name, d.Template, d.Host, d.Port, d.ModbusID, d.Username, d.Password, id)
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

	log.Println("Server listening on :8080")
	handler := enableCORS(mux)
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
