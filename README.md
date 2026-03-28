# Pulse EMS (formerly NEMS)

Pulse EMS is a lightweight, highly responsive, fully UI-driven Energy Management System (EMS) optimized for running on a Raspberry Pi. It provides real-time monitoring, historical tracking, and active control of your home energy usage—coordinating Grid, Solar, Battery, and EV Charger hardware to optimize energy consumption and reduce costs.

## Features

- **Hardware Target**: Deeply optimized for Raspberry Pi (Debian/Linux ARM64).
- **Architecture**: A cohesive monolith featuring a Go (Golang) backend acting as a unified API and web server, paired with a modern Vue 3 Single Page Application (SPA).
- **Minimal SD Card Wear**: Utilizes a highly tuned local SQLite database configured with WAL mode and batched, in-memory transactional writes.
- **Fully UI-Driven**: Zero YAML configuration required. Add, configure, and remove hardware devices entirely through an intuitive frontend.
- **Dynamic Optimization Strategies**:
  - *Eco Mode*: Maximizes self-consumption of solar energy.
  - *Flanders Mode*: Peak shaving to avoid high capacity tariffs.
  - *Netherlands Mode*: Zero-export constraint logic to limit solar feed-in to the grid.
- **Dynamic Pricing**: Integrates Day-Ahead EPEX spot prices for smart scheduling of EV charging and home battery operation.
- **Supported Device Integrations**: Modbus TCP and REST API support for various manufacturers, including Huawei, Raedian, Solis, SMA, Alfen, Easee, and more.

---

## Project Structure

The repository is organized into a monorepo structure:

- `backend/`: Contains the Go application.
  - `main.go`: Entry point, HTTP server setup, and SQLite initialization.
  - `poller.go`: Device data acquisition via `DevicePoller` interfaces and background synchronization.
  - `state.go`: SSE (Server-Sent Events) live streaming and time-series history aggregation.
  - `strategy.go`: Core logic loop for evaluating and executing energy optimization rules.
  - `internal/`: Domain models and hardware-specific device integration templates.
- `frontend/`: Contains the Vue 3 / TypeScript UI.
  - `src/components/`: Core UI cards, the `PowerFlow` hero graphic, and device configuration forms.
  - `src/types/`: Shared TypeScript interface definitions.
- `docs/`: Architectural documentation.

---

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

### Deployment (from tar.gz)

1. **Extract the archive on your Raspberry Pi:**
   ```bash
   mkdir -p /opt/nems
   tar -xzvf build/nems-release-arm64.tar.gz -C /opt/nems
   ```

2. **Set up the systemd service:**
   ```bash
   sudo cp nems.service /etc/systemd/system/
   ```

3. **Create the restricted user environment (recommended):**
   ```bash
   sudo useradd -r -s /bin/false nems
   sudo chown -R nems:nems /opt/nems
   ```

4. **Enable and Start:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable nems.service
   sudo systemctl start nems.service
   ```

---

## Usage

Once the Pulse EMS service is running, it automatically serves both the JSON API and the frontend SPA on port `8080`.

1. **Access the Web UI:**
   Open a web browser on any device in your local network and navigate to:
   ```
   http://<raspberry-pi-ip-address>:8080
   ```

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
