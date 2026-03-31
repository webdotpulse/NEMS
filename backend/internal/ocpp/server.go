package ocpp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for the chargers
	},
	Subprotocols: []string{"ocpp1.6", "ocpp2.0.1"},
}

// GetDeviceProxyUrl is injected from main/poller to retrieve proxy url by host
var GetDeviceProxyUrl func(chargePointId string) string

// OcppState holds the current known state of an OCPP charger
type OcppState struct {
	ChargePointId string
	PowerW        float64
	EnergyWh      float64
	Conn          *websocket.Conn
	mu            sync.RWMutex
	writeMu       sync.Mutex // Protects concurrent writes to Conn
	LastSeen      time.Time

	ProxyConn     *websocket.Conn
	proxyWriteMu  sync.Mutex // Protects concurrent writes to ProxyConn
	OcppProxyUrl  string
}

var (
	chargers   = make(map[string]*OcppState)
	chargersMu sync.RWMutex
)

// GetChargerState retrieves the current state for a given chargePointId
func GetChargerState(chargePointId string) *OcppState {
	chargersMu.RLock()
	defer chargersMu.RUnlock()
	return chargers[chargePointId]
}

// IsConnected safely returns whether the charger currently has an active WebSocket connection
func (s *OcppState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Conn != nil
}

// GetTelemetry safely returns the current power, energy, and last seen time
func (s *OcppState) GetTelemetry() (powerW float64, energyWh float64, lastSeen time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PowerW, s.EnergyWh, s.LastSeen
}

// SendMessage safely writes a JSON message to the WebSocket connection
func (s *OcppState) SendMessage(message []interface{}) error {
	s.mu.RLock()
	conn := s.Conn
	s.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("charger offline")
	}

	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return conn.WriteMessage(websocket.TextMessage, b)
}

// HandleWebSocket handles incoming OCPP WebSocket connections from EV chargers.
// Endpoint typically: /api/ocpp/{chargePointId}
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract chargePointId from URL path (e.g., /api/ocpp/CS-001)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid OCPP endpoint URL", http.StatusBadRequest)
		return
	}
	chargePointId := parts[3]

	// Handle Sec-WebSocket-Protocol (OCPP1.6 or OCPP2.0.1)
	subprotocols := websocket.Subprotocols(r)
	log.Printf("[INFO] OCPP connection request from %s for CP: %s, Subprotocols: %v", r.RemoteAddr, chargePointId, subprotocols)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] OCPP upgrade failed: %v", err)
		return
	}

	log.Printf("[INFO] OCPP CP %s connected via %s", chargePointId, conn.Subprotocol())

	var proxyUrl string
	if GetDeviceProxyUrl != nil {
		proxyUrl = GetDeviceProxyUrl(chargePointId)
	}

	state := &OcppState{
		ChargePointId: chargePointId,
		Conn:          conn,
		LastSeen:      time.Now(),
		OcppProxyUrl:  proxyUrl,
	}

	if proxyUrl != "" {
		log.Printf("[INFO] OCPP CP %s attempting to proxy to %s", chargePointId, proxyUrl)

		dialer := websocket.DefaultDialer
		header := http.Header{}
		if conn.Subprotocol() != "" {
			header.Add("Sec-WebSocket-Protocol", conn.Subprotocol())
		}

		proxyConn, _, err := dialer.Dial(proxyUrl, header)
		if err != nil {
			log.Printf("[WARN] OCPP CP %s proxy connection failed: %v", chargePointId, err)
			state.OcppProxyUrl = "" // Disable proxying for this session
		} else {
			state.ProxyConn = proxyConn
			log.Printf("[INFO] OCPP CP %s proxy connected successfully", chargePointId)

			// Start proxy -> charger forwarder
			go func() {
				defer proxyConn.Close()
				for {
					msgType, msg, err := proxyConn.ReadMessage()
					if err != nil {
						log.Printf("[WARN] OCPP proxy read error for %s: %v", chargePointId, err)
						break
					}
					if msgType == websocket.TextMessage {
						state.writeMu.Lock()
						err = state.Conn.WriteMessage(websocket.TextMessage, msg)
						state.writeMu.Unlock()
						if err != nil {
							log.Printf("[WARN] OCPP write error to charger %s from proxy: %v", chargePointId, err)
							break
						}
					}
				}
				// If proxy disconnects, we might want to close the local connection too,
				// but requirements say "allow the local connection to remain (optionally configurable)".
				// Actually requirement 5: "Ensure both connections close if either side terminates."
				conn.Close()
			}()
		}
	}

	chargersMu.Lock()
	chargers[chargePointId] = state
	chargersMu.Unlock()

	defer func() {
		log.Printf("[INFO] OCPP CP %s disconnected", chargePointId)
		conn.Close()
		if state.ProxyConn != nil {
			state.ProxyConn.Close()
		}
		chargersMu.Lock()
		if chargers[chargePointId] == state {
			// Only remove if it hasn't reconnected and overwritten the map
			state.mu.Lock()
			state.Conn = nil
			state.ProxyConn = nil
			state.mu.Unlock()
		}
		chargersMu.Unlock()
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[ERROR] OCPP read error from %s: %v", chargePointId, err)
			break
		}
		if messageType == websocket.TextMessage {
			state.mu.Lock()
			state.LastSeen = time.Now()
			state.mu.Unlock()

			// Always parse to keep local NEMS state up to date
			handleOcppMessage(conn, state, message, state.ProxyConn != nil)

			// If proxied, forward the raw message to the upstream CSMS
			if state.ProxyConn != nil {
				state.proxyWriteMu.Lock()
				err = state.ProxyConn.WriteMessage(websocket.TextMessage, message)
				state.proxyWriteMu.Unlock()
				if err != nil {
					log.Printf("[WARN] OCPP proxy write error for %s: %v", chargePointId, err)
					// Requirement 5: Ensure both connections close if either side terminates
					conn.Close()
					break
				}
			}
		}
	}
}

// handleOcppMessage parses basic OCPP JSON RPC messages.
// Format: [MessageTypeId, "UniqueId", "Action", {Payload}]
func handleOcppMessage(conn *websocket.Conn, state *OcppState, message []byte, isProxied bool) {
	var raw []interface{}
	if err := json.Unmarshal(message, &raw); err != nil {
		log.Printf("[WARN] OCPP CP %s invalid JSON: %v", state.ChargePointId, err)
		return
	}

	if len(raw) < 3 {
		return
	}

	msgTypeId, ok1 := raw[0].(float64)
	uniqueId, ok2 := raw[1].(string)

	if !ok1 || !ok2 {
		return
	}

	// MessageType 2 is Call (Request)
	if msgTypeId == 2 {
		if len(raw) < 4 {
			return
		}
		action, ok3 := raw[2].(string)
		payload, ok4 := raw[3].(map[string]interface{})

		if !ok3 || !ok4 {
			return
		}

		switch action {
		case "BootNotification":
			if !isProxied {
				// Simple CallResult (Type 3)
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{
						"status":      "Accepted",
						"currentTime": time.Now().Format(time.RFC3339),
						"interval":    300,
					},
				}
				state.SendMessage(response)
			}

		case "Heartbeat":
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{
						"currentTime": time.Now().Format(time.RFC3339),
					},
				}
				state.SendMessage(response)
			}

		case "MeterValues":
			// Parse MeterValues payload to update state
			if meterValue, ok := payload["meterValue"].([]interface{}); ok {
				for _, mv := range meterValue {
					if mvMap, ok := mv.(map[string]interface{}); ok {
						if sampledValue, ok := mvMap["sampledValue"].([]interface{}); ok {
							for _, sv := range sampledValue {
								if svMap, ok := sv.(map[string]interface{}); ok {
									measurand, _ := svMap["measurand"].(string)
									valueStr, _ := svMap["value"].(string)

									var val float64
									fmt.Sscanf(valueStr, "%f", &val)

									state.mu.Lock()
									if measurand == "Power.Active.Import" {
										// Power in Watts
										state.PowerW = val
									} else if measurand == "Energy.Active.Import.Register" || measurand == "" {
										// Energy in Wh
										state.EnergyWh = val
									}
									state.mu.Unlock()
								}
							}
						}
					}
				}
			}
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{},
				}
				state.SendMessage(response)
			}

		case "StatusNotification":
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{},
				}
				state.SendMessage(response)
			}

		case "Authorize":
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{
						"idTagInfo": map[string]interface{}{
							"status": "Accepted",
						},
					},
				}
				state.SendMessage(response)
			}

		case "StartTransaction":
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{
						"idTagInfo": map[string]interface{}{
							"status": "Accepted",
						},
						"transactionId": int(time.Now().Unix()),
					},
				}
				state.SendMessage(response)
			}

		case "StopTransaction":
			if !isProxied {
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{
						"idTagInfo": map[string]interface{}{
							"status": "Accepted",
						},
					},
				}
				state.SendMessage(response)
			}

			// Reset power when transaction stops
			state.mu.Lock()
			state.PowerW = 0
			state.mu.Unlock()

		default:
			log.Printf("[DEBUG] OCPP CP %s Action unhandled: %s", state.ChargePointId, action)
			if !isProxied {
				// Return a generic empty response to prevent charger timeouts
				response := []interface{}{
					3,
					uniqueId,
					map[string]interface{}{},
				}
				state.SendMessage(response)
			}
		}
	}
}

// SetChargingProfile sends a SetChargingProfile request to limit the current.
func (s *OcppState) SetChargingProfile(amps float64) error {
	uniqueId := fmt.Sprintf("%d", time.Now().UnixNano())

	// Basic OCPP 1.6 TxDefaultProfile payload for charging current limit
	payload := map[string]interface{}{
		"connectorId": 0, // 0 = entire charge point
		"csChargingProfile": map[string]interface{}{
			"chargingProfileId":      1,
			"stackLevel":             0,
			"chargingProfilePurpose": "TxDefaultProfile",
			"chargingProfileKind":    "Relative",
			"chargingSchedule": map[string]interface{}{
				"chargingRateUnit": "A",
				"chargingSchedulePeriod": []map[string]interface{}{
					{
						"startPeriod": 0,
						"limit":       amps,
					},
				},
			},
		},
	}

	req := []interface{}{
		2, // Call
		uniqueId,
		"SetChargingProfile",
		payload,
	}

	err := s.SendMessage(req)
	if err != nil {
		log.Printf("[ERROR] OCPP CP %s SetChargingProfile failed: %v", s.ChargePointId, err)
		return err
	}

	log.Printf("[INFO] OCPP CP %s Sent SetChargingProfile limit %.1f A", s.ChargePointId, amps)
	return nil
}
