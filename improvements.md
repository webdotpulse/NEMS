# Proposed Improvements for Pulse EMS

## 1. Real-time Tariff Integration and Automation
Currently, the system automatically fetches Day-Ahead prices using a local provider API (EnergyZero). This could be further enhanced by:
*   Supporting additional local providers natively for better geographical coverage or failovers.

## 2. Enhanced Battery Health and Wear Management
The current arbitrage logic forces charging and discharging based on price spreads. To maximize the lifespan of the home battery, the EMS should track and manage battery wear:
*   Implement cycle counting and Depth of Discharge (DoD) tracking.
*   Add logic to avoid micro-cycling around SOC limits (e.g., oscillating between 99% and 100%).
*   Allow users to configure a "Wear Cost" (in €/kWh) that is factored into the arbitrage spread calculation, ensuring that a discharge cycle is only initiated if the profit exceeds the degradation cost of the battery.

## 3. Multi-zone Temperature Control & Flexible Load Integration
Expand beyond EV chargers and batteries to control flexible thermal loads:
*   Integrate with smart thermostats (e.g., Tado, Nest) and heat pumps (via Modbus/API).
*   Use these systems as thermal batteries: precooling or preheating the house during cheap hours or periods of high solar excess, and coasting through expensive peak hours.

## 4. Machine Learning Based Load Forecasting
The `dynamic_forecast` strategy currently relies on a static user-configured `home_base_load_w` parameter.
*   Replace or supplement this with an adaptive machine learning model that analyzes historical consumption patterns (from the `state.go` time-series data) to predict future home load profiles.
*   This would significantly improve the accuracy of the 24-hour optimal charging plan by anticipating periods of high baseline consumption (e.g., cooking in the evening).

## 5. Proactive Notification and Alerting System
Currently, the system operates autonomously but lacks a direct communication channel to the user for critical events.
*   Implement a notification engine (supporting email, webhooks, or push notifications).
*   Alert users about critical events, such as:
    *   A device going offline (failing to poll).
    *   The projected peak capacity (`ProjectedQuarterPeakW` in Flanders mode) nearing the contract limit despite throttling efforts.
    *   Failing to fetch tariff data.
