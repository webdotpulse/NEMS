package main

import (
	"time"

	"nems/internal/models"
)

// CalculateEffectivePrice returns the effective price per kWh for consumption or injection at a given timestamp
// based on the configured energy contract.
func CalculateEffectivePrice(timestamp time.Time, rawEpexPrice float64, settings models.SiteSettings, isInjection bool) float64 {
	switch settings.ContractType {
	case "fixed":
		if isInjection {
			return settings.EnergyPricesInjection + settings.GridCostsInjection
		}
		return settings.EnergyPricesConsumption + settings.GridCostsConsumption

	case "dynamic":
		if isInjection {
			return (rawEpexPrice * settings.ScaleFactorEpexSpotInjection) + settings.EnergyPricesInjection + settings.GridCostsInjection
		}
		return (rawEpexPrice * settings.ScaleFactorEpexSpotConsumption) + settings.EnergyPricesConsumption + settings.GridCostsConsumption

	case "engie_flextime":
		// Engie EMPOWER Variabel met Flextime
		// Super-dal: 01:00 - 07:00
		// Piek: 07:00 - 11:00 and 17:00 - 22:00
		// Dal: 11:00 - 17:00 and 22:00 - 01:00 (also weekend)
		weekday := timestamp.Weekday()
		hour := timestamp.Hour()

		isWeekend := (weekday == time.Saturday || weekday == time.Sunday)

		markup := 0.0

		if !isWeekend {
			if hour >= 1 && hour < 7 {
				markup = settings.EngieMarkupSuperOffPeak
			} else if (hour >= 7 && hour < 11) || (hour >= 17 && hour < 22) {
				markup = settings.EngieMarkupPeak
			} else {
				markup = settings.EngieMarkupOffPeak
			}
		} else {
			// In the weekend, peak hours are often considered off-peak
			// Based on the prompt: "In het weekend gelden piekuren vaak als daluren"
			// And standard is Dal: weekend 00:00 - 23:59 except super-dal might apply?
			// Prompt says "Daluren: ... (In het weekend gelden piekuren vaak als daluren)."
			// Let's assume standard off-peak for entire weekend, unless 01:00 - 07:00 is super-dal.
			// Let's assume weekend is just off-peak (daluren).
			markup = settings.EngieMarkupOffPeak
		}

		return settings.EngieBaseFee + (rawEpexPrice * settings.EngieMultiplier) + markup

	default:
		// Fallback to dynamic if unknown
		if isInjection {
			return (rawEpexPrice * settings.ScaleFactorEpexSpotInjection) + settings.EnergyPricesInjection + settings.GridCostsInjection
		}
		return (rawEpexPrice * settings.ScaleFactorEpexSpotConsumption) + settings.EnergyPricesConsumption + settings.GridCostsConsumption
	}
}
