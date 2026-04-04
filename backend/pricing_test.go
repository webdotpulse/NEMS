package main

import (
	"fmt"
	"testing"
	"time"

	"nems/internal/models"
)

func TestCalculateEffectivePrice(t *testing.T) {
	// Let's create a base setting for each
	settingsFixed := models.SiteSettings{
		ContractType:         "fixed",
		FixedPricePeakKwh:    0.35,
		FixedPriceOffPeakKwh: 0.25,
	}

	settingsDynamic := models.SiteSettings{
		ContractType:     "dynamic",
		DynamicMarkupKwh: 0.15,
	}

	settingsEngie := models.SiteSettings{
		ContractType:            "engie_flextime",
		EngieBaseFee:            0.05,
		EngieMultiplier:         1.2,
		EngieMarkupSuperOffPeak: 0.01,
		EngieMarkupOffPeak:      0.02,
		EngieMarkupPeak:         0.03,
	}

	rawEpex := 0.10

	// Test Fixed: Monday 10:00 (Peak)
	tMon10 := time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC) // Monday
	if CalculateEffectivePrice(tMon10, rawEpex, settingsFixed) != 0.35 {
		t.Errorf("Fixed peak failed")
	}

	// Test Fixed: Sunday 10:00 (Off-Peak)
	tSun10 := time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC) // Sunday
	if CalculateEffectivePrice(tSun10, rawEpex, settingsFixed) != 0.25 {
		t.Errorf("Fixed off-peak failed")
	}

	// Test Dynamic
	if CalculateEffectivePrice(tMon10, rawEpex, settingsDynamic) != 0.25 {
		t.Errorf("Dynamic failed")
	}

	// Test Engie Super-Dal: Mon 03:00
	tMon03 := time.Date(2023, 10, 2, 3, 0, 0, 0, time.UTC)
	// Base (0.05) + (Raw (0.10) * Mul (1.2)) + SuperOffPeak (0.01) = 0.05 + 0.12 + 0.01 = 0.18
	if fmt.Sprintf("%.2f", CalculateEffectivePrice(tMon03, rawEpex, settingsEngie)) != "0.18" {
		t.Errorf("Engie super dal failed: %v", CalculateEffectivePrice(tMon03, rawEpex, settingsEngie))
	}

	// Test Engie Peak: Mon 08:00
	tMon08 := time.Date(2023, 10, 2, 8, 0, 0, 0, time.UTC)
	// Base (0.05) + (Raw (0.10) * Mul (1.2)) + Peak (0.03) = 0.05 + 0.12 + 0.03 = 0.20
	if fmt.Sprintf("%.2f", CalculateEffectivePrice(tMon08, rawEpex, settingsEngie)) != "0.20" {
		t.Errorf("Engie peak failed: %v", CalculateEffectivePrice(tMon08, rawEpex, settingsEngie))
	}

	// Test Engie Weekend Dal: Sun 08:00
	tSun10 = time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC) // Sunday
	// Base (0.05) + (Raw (0.10) * Mul (1.2)) + OffPeak (0.02) = 0.05 + 0.12 + 0.02 = 0.19
	if fmt.Sprintf("%.2f", CalculateEffectivePrice(tSun10, rawEpex, settingsEngie)) != "0.19" {
		t.Errorf("Engie weekend dal failed: %v", CalculateEffectivePrice(tSun10, rawEpex, settingsEngie))
	}
}
