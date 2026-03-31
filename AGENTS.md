# AGENTS.md - System Directives for EMS Project

## Project Identity
You are building a lightweight, highly responsive, fully UI-driven Energy Management System (EMS) optimized for a Raspberry Pi. The visual language, UI components, and responsiveness should be clean, modern, dark/light mode, and card-based.

## Core Constraints & Rules
1. **NO Car Integration:** Do not include any APIs, UI elements, or logic for vehicle APIs, SOC tracking for cars, or car discovery.
2. **NO Billing/Tokens:** Exclude all funding, payment, RFID token, or billing features.
3. **Fully UI-Driven:** This EMS must allow device configuration, addition, and removal entirely through the frontend UI.
4. **Hardware Target:** Optimized for Raspberry Pi (Debian/Linux). Minimize SD card writes (batch DB writes, use Write-Ahead Logging).

## Required Devices (Templates)
Implement a modular plugin system for devices. The system MUST include templates for:
- Modbus TCP/REST Chargers
- Huawei/SMA/Solis/Enerlution Inverters
- P1/Modbus Smart Meters
- Smart Relays

## UI/UX Requirements
1. **Interactive Energy Flow Chart:** The hero element of the dashboard. Show elements (Grid, Solar, Battery, Charger) ONLY if configured and active.
2. **Click-to-Reveal History:** Clicking any node on the flow chart opens a dedicated view/modal showing historical charts and detailed metrics for *that specific element*.

## Tech Stack
- **Backend:** Go (Golang) or Node.js (TypeScript). Choose the one that ensures the lowest memory footprint on a Raspberry Pi. 
- **Database:** SQLite (with Timeseries capabilities) or minimal InfluxDB for historical data.
- **Frontend:** Vue 3 or React (TypeScript), styled with Tailwind CSS. Use Apache ECharts or Chart.js for historical graphing.
