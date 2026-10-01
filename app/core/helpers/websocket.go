package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
)

var socketSession = struct {
	sync.Mutex
	cancel context.CancelFunc
	conn   *websocket.Conn
}{}

var chatMessageHandler func(WebSocketMessage)

func SetChatMessageHandler(handler func(WebSocketMessage)) {
	chatMessageHandler = handler
}

type WebSocketMessage struct {
	Type           string `json:"type"`
	TargetDeviceID string `json:"targetDeviceID"`
	SenderDeviceID string `json:"senderDeviceID"`
	Role           string `json:"role"`
	Data           any    `json:"data"`
}

// StartWebSocket starts one authenticated, self-reconnecting connection for
// this application instance. HTTP remains the API transport; this is only the
// server-to-client event channel.
func StartWebSocket() error {
	deviceID, err := GetDeviceID()
	if err != nil {
		return err
	}
	if deviceID == "" || GetToken() == "" {
		return nil
	}

	socketSession.Lock()
	if socketSession.cancel != nil {
		socketSession.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	socketSession.cancel = cancel
	socketSession.Unlock()

	go runWebSocket(ctx, deviceID)
	return nil
}

// StopWebSocket ends the event channel immediately, for example during logout.
func StopWebSocket() {
	socketSession.Lock()
	cancel := socketSession.cancel
	socketSession.cancel = nil
	socketSession.Unlock()
	if cancel != nil {
		log.Println("Stopping WebSocket...")
		cancel()
	}
}

func runWebSocket(ctx context.Context, deviceID string) {
	backoff := time.Second

	for ctx.Err() == nil {
		log.Println("Tentando conectar WebSocket...")

		conn, err := dialWebSocket(ctx, deviceID)
		if err == nil {
			log.Println("WebSocket conectado!")

			socketSession.Lock()
			socketSession.conn = conn
			socketSession.Unlock()

			backoff = time.Second

			err = keepWebSocketAlive(ctx, conn)

			socketSession.Lock()
			socketSession.conn = nil
			socketSession.Unlock()

			_ = conn.Close(websocket.StatusNormalClosure, "client stopped")
		}

		if ctx.Err() != nil {
			return
		}

		log.Printf(
			"WebSocket disconnected: %v; reconnecting in %s",
			err,
			backoff,
		)

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func dialWebSocket(ctx context.Context, deviceID string) (*websocket.Conn, error) {
	base, err := url.Parse(apiURL("/ws"))
	if err != nil {
		return nil, err
	}
	switch base.Scheme {
	case "http":
		base.Scheme = "ws"
	case "https":
		base.Scheme = "wss"
	}
	query := base.Query()
	query.Set("device_id", deviceID)
	base.RawQuery = query.Encode()

	header := make(http.Header)
	header.Set("Cookie", "token="+GetToken())
	conn, _, err := websocket.Dial(ctx, base.String(), &websocket.DialOptions{HTTPHeader: header})
	return conn, err
}

func keepWebSocketAlive(ctx context.Context, conn *websocket.Conn) error {
	readErrors := make(chan error, 1)
	go func() {
		for {
			_, message, err := conn.Read(ctx)
			if err != nil {
				readErrors <- err
				return
			}

			handleWebSocketMessage(message)

		}
	}()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-readErrors:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return err
			}
		}
	}
}

func SendWebSocketMessage(
	targetDeviceID string,
	messageType string,
	role string,
	data any,
) error {
	socketSession.Lock()
	conn := socketSession.conn
	socketSession.Unlock()

	if conn == nil {
		return fmt.Errorf("websocket is not connected")
	}

	message := WebSocketMessage{
		Type:           messageType,
		TargetDeviceID: targetDeviceID,
		Role:           role,
		Data:           data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = conn.Write(ctx, websocket.MessageText, payload)

	if err != nil {
		log.Printf(
			"WS SEND ERRO: type=%s target=%s role=%s err=%v",
			messageType,
			targetDeviceID,
			role,
			err,
		)
		return err
	}

	log.Printf(
		"WS SEND OK: type=%s target=%s role=%s",
		messageType,
		targetDeviceID,
		role,
	)

	return nil
}

func handleWebSocketMessage(message []byte) {
	var wsMessage WebSocketMessage

	if err := json.Unmarshal(message, &wsMessage); err != nil {
		log.Println("Invalid WebSocket message:", err)
		return
	}

	log.Printf(
		"WS recebido: type=%s role=%s from=%s",
		wsMessage.Type,
		wsMessage.Role,
		wsMessage.SenderDeviceID,
	)

	switch wsMessage.Type {

	case "chat":
		log.Printf(
			"MESSAGE | From: %s | Data: %v",
			wsMessage.SenderDeviceID,
			wsMessage.Data,
		)

		if chatMessageHandler != nil {
			chatMessageHandler(wsMessage)
		}

	case "webrtc_offer":
		if err := HandleWebRTCOffer(wsMessage); err != nil {
			log.Println("Erro ao tratar WebRTC offer:", err)
		}

	case "webrtc_answer":
		if err := HandleWebRTCAnswer(wsMessage); err != nil {
			log.Println("Erro ao tratar WebRTC answer:", err)
		}

	case "webrtc_ice":
		if err := HandleWebRTCICE(wsMessage); err != nil {
			log.Println("Erro ao tratar WebRTC ICE:", err)
		}

	default:
		log.Println("Unknown WebSocket message:", wsMessage.Type)
	}
}
