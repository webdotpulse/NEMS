# NEMS - Energy Management System

NEMS is a lightweight, highly responsive, fully UI-driven Energy Management System (EMS) optimized for a Raspberry Pi. It provides real-time monitoring and historical tracking of your home energy usage, including Grid, Solar, Battery, and EV Charger metrics.

## Features
- **Hardware Target**: Optimized for Raspberry Pi (Debian/Linux ARM64).
- **Backend**: Go (Golang) server acting as a unified API and web server.
- **Frontend**: Vue 3 SPA built with Tailwind CSS, mimicking a clean, modern, card-based interface.
- **Database**: SQLite optimized for minimal SD card writes (WAL mode).
- **Fully UI-Driven**: No YAML configuration. Add, configure, and remove devices entirely through the frontend.
- **Supported Devices**:
  - EV Charger: Raedian
  - Inverter: Huawei Hybrid Inverter
  - Meter: Huawei Dongle Power Sensor

---

## Installation

### Prerequisites
- A Raspberry Pi (or compatible ARM64 device) running Debian/Linux.
- Node.js (for building the frontend).
- Go (Golang) (for building the backend, requires `CGO_ENABLED=1` and ARM64 C cross-compiler if building from a non-ARM64 machine).

### Building from Source

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd nems
   ```

2. **Run the Build Script:**
   The provided `build.sh` script will install frontend dependencies, compile the Vue SPA, cross-compile the Go backend for `linux/arm64`, and bundle them into a release archive.
   ```bash
   ./build.sh
   ```
   *Output*: `build/nems-release-arm64.tar.gz`

### Installation via DEB Package

If you have built the `.deb` package (e.g., using GitHub Actions or `test_deb.sh`), you can install it directly on your Raspberry Pi:

1. **Copy the `.deb` file to your Raspberry Pi.**
2. **Install the package:**
   ```bash
   sudo dpkg -i nems_1.0.0_arm64.deb
   ```
3. The package automatically creates a `nems` user, sets up the systemd service (`nems.service`), and starts the application.

### Manual Deployment (from tar.gz)

1. **Extract the archive on your Raspberry Pi:**
   ```bash
   mkdir -p /opt/nems
   tar -xzvf nems-release-arm64.tar.gz -C /opt/nems
   ```

2. **Set up the systemd service:**
   Copy the `nems.service` file to the systemd directory:
   ```bash
   sudo cp nems.service /etc/systemd/system/
   ```

3. **Create the NEMS user (recommended):**
   ```bash
   sudo useradd -r -s /bin/false nems
   sudo chown -R nems:nems /opt/nems
   ```

4. **Enable and Start the Service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable nems.service
   sudo systemctl start nems.service
   ```

---

## Usage

Once the NEMS service is running, the application serves both the API and the frontend SPA on port `8080`.

1. **Access the Web UI:**
   Open a web browser and navigate to:
   ```
   http://<raspberry-pi-ip-address>:8080
   ```

2. **Device Configuration:**
   - Go to the **Settings** or **Devices** section in the UI.
   - Click "Add Device" and select a template (Huawei Inverter, Huawei Dongle, or Raedian EV Charger).
   - Enter the device's IP address (Host), Port, and Modbus ID.
   - The system will immediately begin polling the newly added devices and displaying real-time metrics on the dashboard.

3. **Dashboard & Monitoring:**
   - The dashboard displays real-time aggregated site metrics.
   - Historical power data is automatically recorded and can be viewed as charts within the UI.

## Development

If you wish to run the application locally for development:

**Backend:**
```bash
cd backend
go run main.go poller.go state.go
```
*Listens on port 8080.*

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```
*Listens on port 5173.*
