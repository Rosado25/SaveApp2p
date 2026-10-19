package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

const (
	dataChannelLabel   = "saveapp-files"
	maxBufferedAmount  = 512 * 1024
	bufferLowThreshold = 256 * 1024
	p2pOpenTimeout     = 30 * time.Second
)

type peerSession struct {
	mu                sync.Mutex
	targetDeviceID    string
	initiator         bool
	connection        *webrtc.PeerConnection
	dataChannel       *webrtc.DataChannel
	ready             chan struct{}
	readyOnce         sync.Once
	closed            chan struct{}
	closedOnce        sync.Once
	bufferedAmountLow chan struct{}
	pendingCandidates []webrtc.ICECandidateInit
}

var peerSessions = struct {
	sync.RWMutex
	sessions map[string]*peerSession
}{sessions: make(map[string]*peerSession)}

func newPeerSession(targetDeviceID string, connection *webrtc.PeerConnection, initiator bool) *peerSession {
	return &peerSession{
		targetDeviceID:    targetDeviceID,
		initiator:         initiator,
		connection:        connection,
		ready:             make(chan struct{}),
		closed:            make(chan struct{}),
		bufferedAmountLow: make(chan struct{}, 1),
	}
}

func webRTCConfiguration() webrtc.Configuration {
	return webrtc.Configuration{ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}}
}

func sessionFor(targetDeviceID string) *peerSession {
	peerSessions.RLock()
	defer peerSessions.RUnlock()
	return peerSessions.sessions[targetDeviceID]
}

func storeSession(session *peerSession) {
	peerSessions.Lock()
	peerSessions.sessions[session.targetDeviceID] = session
	peerSessions.Unlock()
}

func removeSession(session *peerSession) {
	peerSessions.Lock()
	if peerSessions.sessions[session.targetDeviceID] == session {
		delete(peerSessions.sessions, session.targetDeviceID)
	}
	peerSessions.Unlock()
}

func closeSession(session *peerSession, reason string) {
	session.closedOnce.Do(func() {
		close(session.closed)
		removeSession(session)
		if session.connection != nil {
			_ = session.connection.Close()
		}
		cancelTransfersForDevice(session.targetDeviceID, reason)
		emitTransferEvent(TransferEvent{DeviceID: session.targetDeviceID, State: TransferStateFailed, Message: reason})
	})
}

func configurePeerConnection(session *peerSession) {
	session.connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
			closeSession(session, "ligação P2P terminada")
		}
	})
	session.connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		role := "receiver"
		if session.initiator {
			role = "sender"
		}
		if err := SendWebRTCMessage(session.targetDeviceID, "webrtc_ice", role, candidate.ToJSON()); err != nil {
			log.Println("Erro ao enviar ICE candidate:", err)
		}
	})
}

func configureDataChannel(session *peerSession, dataChannel *webrtc.DataChannel) {
	session.mu.Lock()
	session.dataChannel = dataChannel
	session.mu.Unlock()
	dataChannel.SetBufferedAmountLowThreshold(bufferLowThreshold)
	dataChannel.OnBufferedAmountLow(func() {
		select {
		case session.bufferedAmountLow <- struct{}{}:
		default:
		}
	})
	dataChannel.OnOpen(func() {
		session.readyOnce.Do(func() { close(session.ready) })
		emitTransferEvent(TransferEvent{DeviceID: session.targetDeviceID, State: TransferStateConnected, Message: "canal P2P ligado"})
	})
	dataChannel.OnClose(func() { closeSession(session, "canal P2P fechado") })
	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) { handleDataChannelMessage(session, message) })
}

// StartWebRTCSender creates a P2P connection to a device UUID.
func StartWebRTCSender(targetDeviceID string) (*webrtc.PeerConnection, error) {
	if targetDeviceID == "" {
		return nil, fmt.Errorf("target device ID is empty")
	}
	if existing := sessionFor(targetDeviceID); existing != nil {
		return existing.connection, nil
	}
	connection, err := webrtc.NewPeerConnection(webRTCConfiguration())
	if err != nil {
		return nil, err
	}
	session := newPeerSession(targetDeviceID, connection, true)
	storeSession(session)
	configurePeerConnection(session)
	dataChannel, err := connection.CreateDataChannel(dataChannelLabel, nil)
	if err != nil {
		closeSession(session, "não foi possível criar o canal P2P")
		return nil, err
	}
	configureDataChannel(session, dataChannel)
	offer, err := connection.CreateOffer(nil)
	if err != nil {
		closeSession(session, "não foi possível criar a oferta P2P")
		return nil, err
	}
	if err := connection.SetLocalDescription(offer); err != nil {
		closeSession(session, "não foi possível configurar a oferta P2P")
		return nil, err
	}
	if err := SendWebRTCMessage(targetDeviceID, "webrtc_offer", "sender", offer); err != nil {
		closeSession(session, "não foi possível enviar a oferta P2P")
		return nil, err
	}
	return connection, nil
}

func waitForDataChannel(session *peerSession) error {
	timer := time.NewTimer(p2pOpenTimeout)
	defer timer.Stop()
	select {
	case <-session.ready:
		return nil
	case <-session.closed:
		return fmt.Errorf("a ligação P2P foi fechada")
	case <-timer.C:
		return fmt.Errorf("a ligação P2P excedeu o tempo de espera")
	}
}

func sendText(session *peerSession, data []byte) error {
	if err := waitForDataChannel(session); err != nil {
		return err
	}
	session.mu.Lock()
	dataChannel := session.dataChannel
	session.mu.Unlock()
	if dataChannel == nil || dataChannel.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not connected")
	}
	return dataChannel.SendText(string(data))
}

func sendBinary(session *peerSession, data []byte) error {
	if err := waitForDataChannel(session); err != nil {
		return err
	}
	session.mu.Lock()
	dataChannel := session.dataChannel
	session.mu.Unlock()
	if dataChannel == nil || dataChannel.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not connected")
	}
	for dataChannel.BufferedAmount() > maxBufferedAmount {
		select {
		case <-session.bufferedAmountLow:
		case <-session.closed:
			return fmt.Errorf("a ligação P2P foi fechada")
		}
	}
	return dataChannel.Send(data)
}

func SendData(data string) error {
	peerSessions.RLock()
	defer peerSessions.RUnlock()
	for _, session := range peerSessions.sessions {
		return sendText(session, []byte(data))
	}
	return fmt.Errorf("data channel is not connected")
}

func SendWebRTCMessage(targetDeviceID, messageType, role string, data any) error {
	return SendWebSocketMessage(targetDeviceID, messageType, role, data)
}

func HandleWebRTCAnswer(message WebSocketMessage) error {
	session := sessionFor(message.SenderDeviceID)
	if session == nil {
		return fmt.Errorf("sender peer connection does not exist")
	}
	var answer webrtc.SessionDescription
	data, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &answer); err != nil {
		return err
	}
	if err := session.connection.SetRemoteDescription(answer); err != nil {
		return err
	}
	return addPendingCandidates(session)
}

func addPendingCandidates(session *peerSession) error {
	session.mu.Lock()
	candidates := session.pendingCandidates
	session.pendingCandidates = nil
	session.mu.Unlock()
	for _, candidate := range candidates {
		if err := session.connection.AddICECandidate(candidate); err != nil {
			return err
		}
	}
	return nil
}

func HandleWebRTCICE(message WebSocketMessage) error {
	if message.Role != "sender" && message.Role != "receiver" {
		return fmt.Errorf("WebRTC ICE com role inválido: %q", message.Role)
	}
	var candidate webrtc.ICECandidateInit
	data, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &candidate); err != nil {
		return err
	}
	session := sessionFor(message.SenderDeviceID)
	if session == nil {
		return fmt.Errorf("peer connection does not exist")
	}
	session.mu.Lock()
	if session.connection.RemoteDescription() == nil {
		session.pendingCandidates = append(session.pendingCandidates, candidate)
		session.mu.Unlock()
		return nil
	}
	session.mu.Unlock()
	return session.connection.AddICECandidate(candidate)
}
