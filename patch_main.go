--- backend/main.go
+++ backend/main.go
@@ -95,6 +95,9 @@
	// Insert default if not exists
	_, _ = db.Exec("INSERT OR IGNORE INTO site_settings (id, strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment) VALUES (1, 'eco', 2.5, 0)")

+	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN force_charge_below_euro REAL DEFAULT 0")
+	_, _ = db.Exec("ALTER TABLE site_settings ADD COLUMN smart_ev_cheapest_hours INTEGER DEFAULT 0")
+
	log.Println("Database schema initialized")

	mux := http.NewServeMux()
@@ -151,16 +154,18 @@
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
-			row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment FROM site_settings WHERE id = 1")
+			row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment, force_charge_below_euro, smart_ev_cheapest_hours FROM site_settings WHERE id = 1")
			var settings models.SiteSettings
-			err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment)
+			err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment, &settings.ForceChargeBelowEuro, &settings.SmartEvCheapestHours)
			if err != nil {
				if err == sql.ErrNoRows {
					// Fallback
					settings = models.SiteSettings{
-						StrategyMode: "eco",
-						CapacityPeakLimitKw: 2.5,
+						StrategyMode:              "eco",
+						CapacityPeakLimitKw:       2.5,
						ActiveInverterCurtailment: false,
+						ForceChargeBelowEuro:      0.0,
+						SmartEvCheapestHours:      0,
					}
					json.NewEncoder(w).Encode(settings)
					return
@@ -176,8 +181,8 @@
				return
			}

-			_, err = db.Exec("UPDATE site_settings SET strategy_mode = ?, capacity_peak_limit_kw = ?, active_inverter_curtailment = ? WHERE id = 1",
-				settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment)
+			_, err = db.Exec("UPDATE site_settings SET strategy_mode = ?, capacity_peak_limit_kw = ?, active_inverter_curtailment = ?, force_charge_below_euro = ?, smart_ev_cheapest_hours = ? WHERE id = 1",
+				settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment, settings.ForceChargeBelowEuro, settings.SmartEvCheapestHours)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
