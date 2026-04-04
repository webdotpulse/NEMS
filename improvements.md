# Proposed Improvements for Energy Arbitrage & Optimization

## 1. Implement Round-Trip Efficiency Loss Arbitrage Filtering
Currently, the system uses static Euro thresholds (`force_charge_below_euro`, `force_discharge_above_euro`) to govern arbitrage. This relies heavily on the user perfectly analyzing the difference between prices in the Day-Ahead Market. However, batteries lose typically 15-20% round-trip efficiency (charging and discharging). The arbitrage algorithm should mathematically incorporate this round-trip efficiency penalty. It should refuse to discharge the battery to the grid unless the spread between the purchased price (at time of charging) and the current EPEX spot injection price outweighs the round trip loss cost.

## 2. Differentiate between Consumption and Injection Contracts
Most Belgian dynamic contracts separate consumption formulas and injection formulas (e.g. Injection is often *pure* Belpex * multiplier with zero base fee). Right now, the `CalculateEffectivePrice` function applies a single base formula to `rawEpexPrice` without necessarily factoring in whether the energy is flowing *to* the house or *from* the house. Future iterations should accept an `isInjection bool` and allow users to provide their Injection Multiplier vs Consumption Multiplier separately.

## 3. Dynamic Forecasting Arbitrage vs. Static Thresholds
Right now, the Smart EV charging does check the cheapest hours, but home batteries only check static prices. NEMS should calculate the optimal charging curve for the home battery by forecasting the highest delta spreads over the coming 24 hours, factoring in the solar generation forecast, and charging the battery intelligently during the lowest price slots specifically sized to cover the home's projected net load during the highest price slots.

## 4. Advanced "Superdal" Detection
Currently, only Engie Empower Flextime has hard-coded 'Superdal' specific rules mapped to times 01:00 to 07:00. This could be genericized so that any contract could configure an array of "Off-Peak" and "Super-Peak" blocks, to remove the hard-coded Engie dependencies in the main logic.