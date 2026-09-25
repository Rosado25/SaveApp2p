package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

var (
	receiverMu                sync.Mutex
	receiverConnection        *webrtc.PeerConnection
	receiverDataChannel       *webrtc.DataChannel
	receiverTargetID          string
	receiverPendingCandidates []webrtc.ICECandidateInit
)

// HandleWebRTCOffer trata um Offer recebido.
func HandleWebRTCOffer(message WebSocketMessage) error {
	receiverMu.Lock()
	defer receiverMu.Unlock()

	receiverTargetID = message.SenderDeviceID

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
		return err
	}

	receiverConnection = peerConnection

	// Quando o outro lado criar o DataChannel,
	// este callback será chamado.
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {

		log.Println("DataChannel recebido:", dataChannel.Label())

		receiverDataChannel = dataChannel

		dataChannel.OnOpen(func() {
			log.Println("WebRTC DataChannel aberto!")
		})

		dataChannel.OnClose(func() {
			log.Println("WebRTC DataChannel fechado")
		})

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf(
				"WebRTC recebeu: %s",
				string(msg.Data),
			)
		})
	})

	// ICE candidate.
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateJSON := candidate.ToJSON()

		err := SendWebRTCMessage(
			receiverTargetID,
			"ice_candidate",
			candidateJSON,
		)

		if err != nil {
			log.Println("Erro ao enviar ICE candidate:", err)
		}
	})

	// Converter Data → SessionDescription.
	var offer webrtc.SessionDescription

	data, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &offer); err != nil {
		return err
	}

	// Aplicar Offer.
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return err
	}

	for _, candidate := range receiverPendingCandidates {
		if err := peerConnection.AddICECandidate(candidate); err != nil {
			log.Println("Erro ao adicionar ICE candidate:", err)
		}
	}

	receiverPendingCandidates = nil

	// Criar Answer.
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Aplicar Answer local.
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	// Enviar Answer.
	if err := SendWebRTCMessage(
		receiverTargetID,
		"webrtc_answer",
		answer,
	); err != nil {
		return err
	}

	log.Println("WebRTC Answer enviado!")

	return nil
}

// HandleReceiverICE trata ICE candidates recebidos.
func HandleReceiverICE(message WebSocketMessage) error {
	receiverMu.Lock()
	defer receiverMu.Unlock()

	if receiverConnection == nil {
		return fmt.Errorf("receiver peer connection does not exist")
	}

	var candidate webrtc.ICECandidateInit

	data, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &candidate); err != nil {
		return err
	}

	return receiverConnection.AddICECandidate(candidate)
}

// SendReceiverData envia dados pelo DataChannel.
func SendReceiverData(data string) error {
	receiverMu.Lock()
	defer receiverMu.Unlock()

	if receiverDataChannel == nil {
		return fmt.Errorf("data channel is not connected")
	}

	return receiverDataChannel.SendText(data)
}
