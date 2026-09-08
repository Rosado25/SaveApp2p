package helpers

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	loginWindow   = 15 * time.Minute
	loginMaxFails = 5
)

type loginAttempt struct {
	failures  int
	startedAt time.Time
}

var loginAttempts = struct {
	sync.Mutex
	clients map[string]loginAttempt
}{clients: make(map[string]loginAttempt)}

func ClientAddress(r *http.Request) string {
	address, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return address
	}
	return r.RemoteAddr
}

func AllowLogin(client string) bool {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()

	attempt, exists := loginAttempts.clients[client]
	if !exists || time.Since(attempt.startedAt) >= loginWindow {
		return true
	}
	return attempt.failures < loginMaxFails
}

func RecordLoginFailure(client string) {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()

	attempt, exists := loginAttempts.clients[client]
	if !exists || time.Since(attempt.startedAt) >= loginWindow {
		attempt = loginAttempt{startedAt: time.Now()}
	}
	attempt.failures++
	loginAttempts.clients[client] = attempt
}

func ResetLoginFailures(client string) {
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	delete(loginAttempts.clients, client)
}
