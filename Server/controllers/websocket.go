package controllers

import (
	"Server/helpers"
	"Server/services"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

// ConnectionHub keeps the currently-online device connections. The hub is
// intentionally independent from HTTP routes so future WebRTC signalling can
// publish an event to a specific connected device.
type ConnectionHub struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
}

func NewConnectionHub() *ConnectionHub {
	return &ConnectionHub{connections: make(map[string]*websocket.Conn)}
}

func (h *ConnectionHub) add(deviceID string, conn *websocket.Conn) (previous *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	previous = h.connections[deviceID]
	h.connections[deviceID] = conn
	return previous
}

func (h *ConnectionHub) remove(deviceID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connections[deviceID] == conn {
		delete(h.connections, deviceID)
	}
}

// SendToDevice is the server-side entry point for future signalling events.
// It returns false when that device is not currently connected.
func (h *ConnectionHub) SendToDevice(ctx context.Context, deviceID string, message []byte) bool {
	h.mu.RLock()
	conn := h.connections[deviceID]
	h.mu.RUnlock()
	if conn == nil {
		return false
	}
	return conn.Write(ctx, websocket.MessageText, message) == nil
}

type WebSocketController struct {
	devices *services.DeviceService
	hub     *ConnectionHub
}

func NewWebSocketController(devices *services.DeviceService, hub *ConnectionHub) *WebSocketController {
	return &WebSocketController{devices: devices, hub: hub}
}

func (c *WebSocketController) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	token, err := helpers.GetTokenFromCookie(r)
	if err != nil {
		helpers.JSON(w, http.StatusUnauthorized, "Not logged in")
		return
	}
	userID, err := helpers.ValidateJWT(token)
	if err != nil {
		helpers.JSON(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	deviceID := strings.TrimSpace(r.URL.Query().Get("device_id"))
	if deviceID == "" || len(deviceID) > 128 {
		helpers.JSON(w, http.StatusBadRequest, "A valid device_id is required")
		return
	}
	belongsToUser, err := c.devices.DeviceBelongsToUser(userID, deviceID)
	if err != nil || !belongsToUser {
		helpers.JSON(w, http.StatusForbidden, "Device is not associated with this user")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Native clients authenticate with the JWT cookie.
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	if previous := c.hub.add(deviceID, conn); previous != nil && previous != conn {
		_ = previous.Close(websocket.StatusPolicyViolation, "replaced by a new connection")
	}
	defer c.hub.remove(deviceID, conn)

	for {
		_, message, err := conn.Read(r.Context())
		if err != nil {
			return
		}

		var wsMessage WebSocketMessage

		if err := json.Unmarshal(message, &wsMessage); err != nil {
			continue
		}

		if wsMessage.TargetDeviceID == "" {
			continue
		}

		targetDevice, err := c.devices.GetDeviceByID(wsMessage.TargetDeviceID)
		if err != nil {
			continue
		}

		wsMessage.SenderDeviceID = deviceID

		message, err = json.Marshal(wsMessage)
		if err != nil {
			continue
		}

		log.Printf(
			"WS encaminhar: type=%s from=%s -> target=%s role=%s",
			wsMessage.Type,
			wsMessage.SenderDeviceID,
			targetDevice.DeviceID,
			wsMessage.Role,
		)

		c.hub.SendToDevice(
			r.Context(),
			targetDevice.DeviceID,
			message,
		)
	}
}

type WebSocketMessage struct {
	Type           string `json:"type"`
	TargetDeviceID string `json:"targetDeviceID"`
	SenderDeviceID string `json:"senderDeviceID"`
	Role           string `json:"role"`
	Data           any    `json:"data"`
}
