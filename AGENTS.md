# AGENTS.md - System Directives for EMS Project

## Project Identity
You are building a lightweight, highly responsive, fully UI-driven Energy Management System (EMS) optimized for a Raspberry Pi. The visual language, UI components, and responsiveness should be clean, modern, dark/light mode, and card-based.

## Core Constraints & Rules
1. **NO Car Integration:** Do not include any APIs, UI elements, or logic for vehicle APIs, SOC tracking for cars, or car discovery.
2. **NO Billing/Tokens:** Exclude all funding, payment, RFID token, or billing features.
3. **NO Imbalance Logic:** Do not implement imbalance logic.
4. **NO ENTSO-E API:** Do not implement the ENTSO-E API.
5. **Target Audience:** Keep this EMS specifically designed for home owners and small installations.
6. **Elia API:** Only include the Elia API if possible or necessary for features other than imbalance.
7. **Fully UI-Driven:** This EMS must allow device configuration, addition, and removal entirely through the frontend UI.
8. **Hardware Target:** Optimized for Raspberry Pi (Debian/Linux). Minimize SD card writes (batch DB writes, use Write-Ahead Logging).
9. **System Info & Build Number:** Do NOT include CPU information in the System Info UI. The build number must use the release tag (`git describe --tags --always`).
10. **Deployment:** Deployed via a `.deb` package that creates a `nems` user and `nems.service`. Also built as a custom Raspberry Pi OS image with `nginx` proxy, a minimal Wayland desktop (`wayfire`), `cockpit` for web-based terminal access on port 9090, and `rpi-connect` for remote screen sharing.

## Required Devices (Templates)
Implement a modular plugin system for devices. The system MUST include templates for:
- OCPP 1.6 / 2.0.1 EV Chargers
- Modbus TCP/REST Chargers
- Huawei/SMA/Solis/Enerlution Inverters
- P1/Modbus Smart Meters
- Smart Relays

## UI/UX Requirements
1. **Interactive Energy Flow Chart:** The hero element of the dashboard. Show elements (Grid, Solar, Battery, Charger) ONLY if configured and active.
2. **Click-to-Reveal History:** Clicking any node on the flow chart opens a dedicated view/modal showing historical charts and detailed metrics for *that specific element*.
3. **Data Exportation:** The system allows exporting of both technical logs (via the Logger UI) and user-friendly visual energy reports (via the Dashboard UI in PDF format).

## Tech Stack
- **Backend:** Go (Golang), chosen to ensure the lowest memory footprint on a Raspberry Pi.
- **Database:** SQLite (with Timeseries capabilities) for historical data.
- **Frontend:** Vue 3 (TypeScript), styled with Tailwind CSS. Uses Chart.js for historical graphing.
