package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

var (
	senderMu                sync.Mutex
	senderConnection        *webrtc.PeerConnection
	senderDataChannel       *webrtc.DataChannel
	senderPendingCandidates []webrtc.ICECandidateInit
)

// StartWebRTCSender cria uma ligação WebRTC para outro device.
func StartWebRTCSender(targetDeviceID string) (*webrtc.PeerConnection, error) {
	senderMu.Lock()
	defer senderMu.Unlock()

	if senderConnection != nil {
		return senderConnection, nil
	}

	// Configuração inicial.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:stun.l.google.com:19302",
				},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	senderConnection = peerConnection

	// Criar DataChannel.
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		peerConnection.Close()
		senderConnection = nil
		return nil, err
	}

	senderDataChannel = dataChannel

	dataChannel.OnOpen(func() {
		log.Println("WebRTC DataChannel aberto!")
	})

	dataChannel.OnClose(func() {
		log.Println("WebRTC DataChannel fechado")
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("WebRTC recebeu: %s", string(msg.Data))
	})

	// ICE candidate.
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateJSON := candidate.ToJSON()

		SendWebRTCMessage(
			targetDeviceID,
			"ice_candidate",
			candidateJSON,
		)
	})

	// Criar Offer.
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		senderConnection = nil
		return nil, err
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		peerConnection.Close()
		senderConnection = nil
		return nil, err
	}

	// Enviar Offer pelo WebSocket.
	if err := SendWebRTCMessage(
		targetDeviceID,
		"webrtc_offer",
		offer,
	); err != nil {
		peerConnection.Close()
		senderConnection = nil
		return nil, err
	}

	log.Println("WebRTC Offer enviado!")

	return peerConnection, nil
}

// SendWebRTCMessage envia uma mensagem WebRTC através do WebSocket.
func SendWebRTCMessage(
	targetDeviceID string,
	messageType string,
	data any,
) error {
	return SendWebSocketMessage(
		targetDeviceID,
		messageType,
		data,
	)
}

// SendData envia dados através do DataChannel.
func SendData(data string) error {
	senderMu.Lock()
	defer senderMu.Unlock()

	if senderDataChannel == nil {
		return fmt.Errorf("data channel is not connected")
	}

	return senderDataChannel.SendText(data)
}

// HandleWebRTCAnswer trata o Answer recebido.
func HandleWebRTCAnswer(message WebSocketMessage) error {
	senderMu.Lock()
	defer senderMu.Unlock()

	if senderConnection == nil {
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

	if err := senderConnection.SetRemoteDescription(answer); err != nil {
		return err
	}

	for _, candidate := range senderPendingCandidates {
		if err := senderConnection.AddICECandidate(candidate); err != nil {
			log.Println("Erro ao adicionar ICE candidate:", err)
		}
	}

	senderPendingCandidates = nil

	log.Println("WebRTC Answer recebido!")

	return nil
}

// HandleWebRTCICE trata ICE candidates recebidos.
func HandleWebRTCICE(message WebSocketMessage) error {
	senderMu.Lock()
	defer senderMu.Unlock()

	if senderConnection == nil {
		return fmt.Errorf("sender peer connection does not exist")
	}

	var candidate webrtc.ICECandidateInit

	data, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &candidate); err != nil {
		return err
	}

	if senderConnection.RemoteDescription() == nil {
		senderPendingCandidates = append(
			senderPendingCandidates,
			candidate,
		)

		log.Println("ICE candidate guardado para depois")

		return nil
	}

	return senderConnection.AddICECandidate(candidate)
}
