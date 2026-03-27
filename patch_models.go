--- backend/internal/models/models.go
+++ backend/internal/models/models.go
@@ -3,6 +3,8 @@
 type SiteSettings struct {
	StrategyMode              string  `json:"strategy_mode"`
	CapacityPeakLimitKw       float64 `json:"capacity_peak_limit_kw"`
	ActiveInverterCurtailment bool    `json:"active_inverter_curtailment"`
+	ForceChargeBelowEuro      float64 `json:"force_charge_below_euro"`
+	SmartEvCheapestHours      int     `json:"smart_ev_cheapest_hours"`
 }

 type Device struct {
