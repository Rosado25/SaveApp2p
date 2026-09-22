package helpers

import (
	"context"
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
}{}

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
		conn, err := dialWebSocket(ctx, deviceID)
		if err == nil {
			backoff = time.Second
			err = keepWebSocketAlive(ctx, conn)
			_ = conn.Close(websocket.StatusNormalClosure, "client stopped")
		}
		if ctx.Err() != nil {
			return
		}
		log.Printf("WebSocket disconnected: %v; reconnecting in %s", err, backoff)
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
			_, _, err := conn.Read(ctx)
			if err != nil {
				readErrors <- err
				return
			}
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
