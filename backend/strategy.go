package main

import (
	"database/sql"
	"log"
	"time"

	"nems/internal/models"
)

type StrategyController struct {
	stopCh chan struct{}
}

var StrategyCtrl *StrategyController

func InitStrategyController() {
	StrategyCtrl = &StrategyController{
		stopCh: make(chan struct{}),
	}
}

func (sc *StrategyController) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var settings models.SiteSettings
				row := db.QueryRow("SELECT strategy_mode, capacity_peak_limit_kw, active_inverter_curtailment FROM site_settings WHERE id = 1")
				err := row.Scan(&settings.StrategyMode, &settings.CapacityPeakLimitKw, &settings.ActiveInverterCurtailment)
				if err != nil {
					if err == sql.ErrNoRows {
						settings = models.SiteSettings{
							StrategyMode: "eco",
							CapacityPeakLimitKw: 2.5,
							ActiveInverterCurtailment: false,
						}
					} else {
						log.Printf("StrategyController: Error fetching site settings: %v", err)
						continue
					}
				}
				log.Printf("StrategyController running - Mode: %s, Peak Limit: %.1f kW, Curtailment: %v",
					settings.StrategyMode, settings.CapacityPeakLimitKw, settings.ActiveInverterCurtailment)
			case <-sc.stopCh:
				log.Println("StrategyController stopped")
				return
			}
		}
	}()
}

func (sc *StrategyController) Stop() {
	close(sc.stopCh)
}
