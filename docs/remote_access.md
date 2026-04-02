# Remote Access Documentation

Pulse EMS custom OS images include two powerful tools to help you manage your Raspberry Pi and access your dashboard remotely: **Cockpit** and **Raspberry Pi Connect**.

## Cockpit

Cockpit is a web-based graphical interface for servers. It allows you to manage system services, monitor performance, and access a terminal directly from your web browser.

### Accessing Cockpit
1. Ensure your Raspberry Pi is connected to your local network and powered on.
2. Open a web browser on a device on the same network.
3. Navigate to:
   ```
   http://<raspberry-pi-ip-address>:9090
   ```
   *(Or `http://ems.local:9090` if your network supports mDNS).*
4. **Login Credentials**:
   - Username: `admin`
   - Password: `manufacturer`
   *(It is highly recommended to change this password after your first login via the terminal or Cockpit interface).*

### Features
- **System Monitoring:** View CPU, Memory, and Network usage.
- **Service Management:** Start, stop, and inspect systemd services (e.g., `nems.service` or `nginx`).
- **Terminal Access:** Access a full root-capable bash shell without needing an SSH client.
- **Log Viewer:** Easily read system logs (`journalctl`) to diagnose hardware or connectivity issues.

---

## Raspberry Pi Connect

Raspberry Pi Connect provides secure remote screen sharing, allowing you to access the minimal desktop environment (and the NEMS dashboard) from anywhere in the world, without setting up a VPN or configuring port forwarding on your router.

The custom OS image includes a minimal Wayland desktop (`wayfire`), `chromium-browser`, and the `rpi-connect` package.

### Initial Setup & Configuration
You must start the service and pair your Raspberry Pi to your Raspberry Pi ID.

1. **Access the Terminal:** You can use SSH or the terminal provided by **Cockpit** (see above).
2. **Start the Service:**
   ```bash
   systemctl --user start rpi-connect
   ```
3. **Enable on Boot (Optional but recommended):**
   ```bash
   systemctl --user enable rpi-connect
   ```
4. **Pair the Device:**
   Run the following command to generate a pairing link:
   ```bash
   rpi-connect signin
   ```
   *Follow the provided URL in your browser to log in to your Raspberry Pi ID and complete the pairing process.*

### Accessing the Dashboard Remotely
Once paired:
1. Go to [connect.raspberrypi.com](https://connect.raspberrypi.com) and log in.
2. Select your device and choose **Screen Sharing**.
3. You will be presented with the minimal desktop. Open the Chromium browser and navigate to `http://127.0.0.1:8080` (or `http://localhost:8080`) to view and manage your Pulse EMS dashboard.