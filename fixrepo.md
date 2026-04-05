# Pulse EMS Improvements and Fixes

## Removed Unused Code
- Completely removed the legacy `OcppProxyUrl` bidirectional proxy functionality from the built-in OCPP server. The EMS is designed for home owners and small installations, so forwarding telemetry to an upstream corporate CSMS is unnecessary and adds complexity.
  - Removed proxy connection logic and state tracking from `backend/internal/ocpp/server.go`.
  - Removed `ocpp_proxy_url` fields from the database schema (`backend/main.go`), API responses (`backend/api.go`), device models (`backend/internal/models/models.go`), and polling logic (`backend/poller.go` and `backend/internal/templates/ocpp.go`).

## Documentation Updates
- Updated `README.md` to remove references to the optional bi-directional proxy.
- Updated `docs/architecture.md` to remove references to the optional bi-directional proxy.

## Bullet-proofing Existing Logic
- Removed dead and unused code paths.
- Ensured all concurrent map accesses (like `chargersMu`) in the OCPP server correctly utilize `sync.RWMutex` locks to prevent race conditions when handling WebSocket messages.
