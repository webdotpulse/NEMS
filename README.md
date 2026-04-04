# Pulse EMS

Pulse EMS is a lightweight, highly responsive, fully UI-driven Energy Management System (EMS) optimized for running on a Raspberry Pi. It provides real-time monitoring, historical tracking, and active control of your home energy usage—coordinating Grid, Solar, Battery, and EV Charger hardware to optimize energy consumption and reduce costs.

## Features

- **Hardware Target**: Deeply optimized for Raspberry Pi (Debian/Linux ARM64). CPU metrics are intentionally excluded from the UI to maintain a lightweight profile. The system displays the release tag as its build number.
- **Architecture**: A cohesive monolith featuring a Go (Golang) backend acting as a unified API and web server, paired with a modern Vue 3 Single Page Application (SPA).
- **Minimal SD Card Wear**: Utilizes a highly tuned local SQLite database configured with WAL mode and batched, in-memory transactional writes.
- **Fully UI-Driven**: Zero YAML configuration required. Add, configure, and remove hardware devices entirely through an intuitive frontend.
- **Dynamic Optimization Strategies**:
  - *Eco Mode*: Maximizes self-consumption of solar energy.
  - *Flanders Mode (Predictive Peak Shaving)*: Uses predictive instantaneous power limit calculations based on elapsed time within a synchronized 15-minute window to actively throttle EV chargers and batteries, safely capping the `ProjectedQuarterPeakW` and avoiding high capacity tariffs.
  - *Netherlands Mode*: Zero-export constraint logic to limit solar feed-in to the grid.
- **Dynamic Battery Arbitrage**: Integrates Day-Ahead EPEX spot prices to allow users to force charge batteries during the $N$ cheapest hours and force discharge during the $M$ most expensive hours, rather than relying strictly on static threshold values.
- **Native OCPP Server & Proxy**: A built-in OCPP 1.6 / 2.0.1 WebSocket server allows EV chargers to connect directly to the EMS. Includes an optional bi-directional proxy to seamlessly forward telemetry to an upstream CSMS (like a corporate backend) while intercepting data for local EMS optimization.
- **Network Scanner**: A zero-dependency network discovery tool leveraging localized MAC OUI maps to instantly find and identify supported hardware on your local network.
- **Supported Device Integrations**: OCPP 1.6 / 2.0.1, Modbus TCP, and REST API support for various manufacturers, including Huawei, Raedian, Solis, SMA, Easee, Enerlution, HomeWizard, and P1 meters.

## Prerequisites & Installation

### Prerequisites

- A Raspberry Pi (or compatible ARM64 device) running Debian/Linux.
- Node.js (for building the frontend assets).
- Go (Golang) (for building the backend).
  - *Note: If building from a non-ARM64 machine, you must enable `CGO_ENABLED=1` and utilize an ARM64 C cross-compiler (e.g., `gcc-aarch64-linux-gnu`) due to the `mattn/go-sqlite3` dependency.*

### Building from Source

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd nems
   ```

2. **Compile the Monolith:**
   The provided script will install frontend dependencies, compile the Vue SPA into the `dist` folder, cross-compile the Go backend for `linux/arm64`, and bundle them into a single archive.
   ```bash
   ./build.sh
   ```
   *Output*: `build/nems-release-arm64.tar.gz`

### Deployment

Pulse EMS is primarily deployed via release artifacts generated automatically by our CI/CD pipeline: a Debian (`.deb`) package and a fully pre-configured custom Raspberry Pi OS image.

#### Option A: Flash the Custom OS Image (Recommended)
The easiest way to get started is to flash our pre-built custom Raspberry Pi OS Lite (ARM64) image onto your SD card. This image is built on Bookworm and comes pre-installed with NEMS, Nginx as a reverse proxy, and remote access tools.
1. Download `nems-os-image.img.xz` from the latest release.
2. Use a tool like BalenaEtcher or Raspberry Pi Imager to flash it to an SD card.
3. Insert the SD card into your Pi and boot.
The host is configured as `ems` (e.g. accessible at `http://ems` or `http://ems.local`).
The image also includes Cockpit pre-installed for web-based terminal access and system management on port 9090. See [Remote Access Documentation](docs/remote_access.md) for details on Cockpit and Raspberry Pi Connect.

#### Option B: Install via Debian Package (.deb)
If you already have a compatible Debian/Ubuntu ARM64 system running:
1. Download the latest `nems_*_arm64.deb` release.
2. Install the package:
   ```bash
   sudo dpkg -i nems_*_arm64.deb
   sudo apt-get install -f # to resolve any dependencies
   ```
The package automatically creates the restricted `nems` user environment, installs the binary and frontend to `/opt/nems`, and configures and starts the `nems.service` via systemd.

### Remote Access (Cockpit & Raspberry Pi Connect)
If you are using the Custom OS Image (Option A), it includes powerful tools for remote management:
- **Cockpit**: A web-based graphical interface for system management, terminal access, and performance monitoring.
- **Raspberry Pi Connect**: Enables secure remote screen sharing to access the NEMS dashboard from anywhere without a VPN.

For detailed setup instructions, please see the [Remote Access Documentation](docs/remote_access.md).

## Usage

Once the Pulse EMS service is running, it automatically serves both the JSON API and the frontend SPA. The service multiplexes both standard HTTP and secure HTTPS traffic on the same ports.
By default, the application binds to port `8080` (primary). If sufficient permissions are available, it also attempts to bind to standard web ports `80` and `443`.

1. **Access the Web UI:**
   Open a web browser on any device in your local network and navigate to:
   ```
   http://<raspberry-pi-ip-address>:8080
   ```
   *Note: You can also use HTTPS (e.g., `https://<raspberry-pi-ip-address>:8080`). If DNS is configured, navigating to `http://ems.local` or `https://ems.local` will also work directly without port specification.*

2. **Initial Setup:**
   - Navigate to the **Settings** view in the top navigation bar.
   - Under the "Configured Devices" section, click to "Add Device".
   - Select the appropriate template for your hardware (e.g., *Huawei Hybrid Inverter* or *Raedian EV Charger*).
   - Enter the device's IP address (Host), Port, Modbus ID, and any relevant credentials.
   - Save the device. The system will immediately begin polling the hardware.

3. **Monitoring Dashboard:**
   - Return to the **Dashboard** to view real-time aggregated metrics via the interactive Power Flow diagram.
   - Daily performance metrics (Grid Import/Export, Solar Yield, etc.) are calculated continuously.

4. **Applying Strategies:**
   - In the **Settings** view, locate the "Site Optimization" section.
   - Select an Optimization Strategy (like *Eco* or *Netherlands Mode*) and configure associated parameters like Grid Nominal Current or Allowed Export limits to enable autonomous hardware control.

## Project Structure

The repository is organized into a monorepo structure:

- `backend/`: Contains the Go application.
  - `main.go`: Entry point, HTTP server setup, and SQLite initialization.
  - `api.go`: Houses the API HTTP handlers.
  - `poller.go`: Device data acquisition via `DevicePoller` interfaces and background synchronization.
  - `state.go`: SSE (Server-Sent Events) live streaming and time-series history aggregation.
  - `strategy.go`: Core logic loop for evaluating and executing energy optimization rules.
  - `internal/`: Domain models and hardware-specific device integration templates.
- `frontend/`: Contains the Vue 3 / TypeScript UI.
  - `src/components/`: Core UI cards, the `PowerFlow` hero graphic, and device configuration forms.
  - `src/types/`: Shared TypeScript interface definitions.
- `docs/`: Architectural documentation.
