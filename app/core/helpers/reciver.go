package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v4"
)

// HandleWebRTCOffer accepts an incoming connection from the verified sender.
func HandleWebRTCOffer(message WebSocketMessage) error {
	if message.SenderDeviceID == "" {
		return fmt.Errorf("sender device ID is empty")
	}
	if previous := sessionFor(message.SenderDeviceID); previous != nil {
		closeSession(previous, "substituída por uma nova ligação P2P")
	}
	connection, err := webrtc.NewPeerConnection(webRTCConfiguration())
	if err != nil {
		return err
	}
	session := newPeerSession(message.SenderDeviceID, connection, false)
	storeSession(session)
	configurePeerConnection(session)
	connection.OnDataChannel(func(dataChannel *webrtc.DataChannel) { configureDataChannel(session, dataChannel) })
	var offer webrtc.SessionDescription
	data, err := json.Marshal(message.Data)
	if err != nil {
		closeSession(session, "oferta P2P inválida")
		return err
	}
	if err := json.Unmarshal(data, &offer); err != nil {
		closeSession(session, "oferta P2P inválida")
		return err
	}
	if err := connection.SetRemoteDescription(offer); err != nil {
		closeSession(session, "não foi possível aceitar a oferta P2P")
		return err
	}
	if err := addPendingCandidates(session); err != nil {
		closeSession(session, "candidato ICE inválido")
		return err
	}
	answer, err := connection.CreateAnswer(nil)
	if err != nil {
		closeSession(session, "não foi possível criar a resposta P2P")
		return err
	}
	if err := connection.SetLocalDescription(answer); err != nil {
		closeSession(session, "não foi possível configurar a resposta P2P")
		return err
	}
	if err := SendWebRTCMessage(message.SenderDeviceID, "webrtc_answer", "receiver", answer); err != nil {
		closeSession(session, "não foi possível enviar a resposta P2P")
		return err
	}
	return nil
}

func HandleReceiverICE(message WebSocketMessage) error { return HandleWebRTCICE(message) }

func SendReceiverData(data string) error { return SendData(data) }
