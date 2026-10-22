package helpers

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

const (
	chunkSize = 16 * 1024
	//Max tranfer size is 16 GiB.
	maxTransferSize = 16 * 1024 * 1024 * 1024
	chunkHeaderSize = 29
)

type TransferState string

const (
	TransferStateConnected TransferState = "connected"
	TransferStateOffered   TransferState = "offered"
	TransferStatePending   TransferState = "pending"
	TransferStateSending   TransferState = "sending"
	TransferStateReceiving TransferState = "receiving"
	TransferStateCompleted TransferState = "completed"
	TransferStateRejected  TransferState = "rejected"
	TransferStateCancelled TransferState = "cancelled"
	TransferStateFailed    TransferState = "failed"
)

type TransferEvent struct {
	ID       string
	DeviceID string
	Name     string
	Size     int64
	Progress int64
	Incoming bool
	State    TransferState
	Message  string
}

type fileOffer struct {
	Type       string `json:"type"`
	TransferID string `json:"transferID"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	MIME       string `json:"mime"`
	SHA256     string `json:"sha256"`
}

type fileControl struct {
	Type       string `json:"type"`
	TransferID string `json:"transferID"`
	Message    string `json:"message,omitempty"`
}

type fileTransfer struct {
	mu            sync.Mutex
	id            string
	deviceID      string
	name          string
	size          int64
	sha256        string
	incoming      bool
	path          string
	temporaryPath string
	file          *os.File
	hash          hashWriter
	nextChunk     uint64
	progress      int64
	lastEvent     time.Time
	cancelled     bool
	pendingTimer  *time.Timer
}

type hashWriter interface {
	Write([]byte) (int, error)
	Sum([]byte) []byte
}

var transferRegistry = struct {
	sync.RWMutex
	transfers map[string]*fileTransfer
	handler   func(TransferEvent)
}{transfers: make(map[string]*fileTransfer)}

func SetTransferEventHandler(handler func(TransferEvent)) {
	transferRegistry.Lock()
	transferRegistry.handler = handler
	transferRegistry.Unlock()
}

func emitTransferEvent(event TransferEvent) {
	transferRegistry.RLock()
	handler := transferRegistry.handler
	transferRegistry.RUnlock()
	if handler != nil {
		handler(event)
	}
}

func registerTransfer(transfer *fileTransfer) {
	transferRegistry.Lock()
	transferRegistry.transfers[transfer.id] = transfer
	transferRegistry.Unlock()
}

func findTransfer(id string) *fileTransfer {
	transferRegistry.RLock()
	defer transferRegistry.RUnlock()
	return transferRegistry.transfers[id]
}

func removeTransfer(id string) {
	transferRegistry.Lock()
	delete(transferRegistry.transfers, id)
	transferRegistry.Unlock()
}

// SendFile starts a transfer after the remote device explicitly accepts it.
func SendFile(targetDeviceID, path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("o caminho selecionado não é um ficheiro regular")
	}
	if info.Size() > maxTransferSize {
		return "", fmt.Errorf("o ficheiro excede o limite de %d GiB", maxTransferSize/(1024*1024*1024))
	}
	transfer := &fileTransfer{
		id:       uuid.NewString(),
		deviceID: targetDeviceID,
		name:     filepath.Base(path),
		size:     info.Size(),
		path:     path,
	}
	registerTransfer(transfer)
	emitTransferEvent(eventFor(transfer, TransferStateOffered, "a ligar ao dispositivo"))

	if _, err := StartWebRTCSender(targetDeviceID); err != nil {
		removeTransfer(transfer.id)
		return "", err
	}
	go prepareAndOfferFile(transfer)
	return transfer.id, nil
}

func prepareAndOfferFile(transfer *fileTransfer) {
	digest, err := hashFile(transfer.path)
	if err != nil {
		failTransfer(transfer, "não foi possível calcular a integridade do ficheiro")
		return
	}
	transfer.mu.Lock()
	if transfer.cancelled {
		transfer.mu.Unlock()
		return
	}
	transfer.sha256 = digest
	transfer.mu.Unlock()
	offerFile(transfer)
}

func offerFile(transfer *fileTransfer) {
	session := sessionFor(transfer.deviceID)
	if session == nil || waitForDataChannel(session) != nil {
		failTransfer(transfer, "não foi possível abrir o canal P2P")
		return
	}
	payload, err := json.Marshal(fileOffer{Type: "file_offer", TransferID: transfer.id, Name: transfer.name, Size: transfer.size, MIME: mime.TypeByExtension(filepath.Ext(transfer.name)), SHA256: transfer.sha256})
	if err != nil || sendText(session, payload) != nil {
		failTransfer(transfer, "não foi possível enviar a proposta do ficheiro")
		return
	}
	setPendingTimeout(transfer)
	emitTransferEvent(eventFor(transfer, TransferStatePending, "a aguardar aceitação"))
}

// AcceptIncomingFile creates a temporary file in directory and authorizes the sender.
func AcceptIncomingFile(transferID, directory string) error {
	transfer := findTransfer(transferID)
	if transfer == nil || !transfer.incoming {
		return errors.New("transferência recebida não encontrada")
	}
	stopPendingTimeout(transfer)
	transfer.mu.Lock()
	defer transfer.mu.Unlock()
	if transfer.file != nil {
		return errors.New("transferência já foi aceite")
	}
	if directory == "" {
		return errors.New("pasta de destino não selecionada")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	destination, err := availableFilePath(directory, transfer.name)
	if err != nil {
		return err
	}
	temporary := destination + ".part"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	transfer.path = destination
	transfer.temporaryPath = temporary
	transfer.file = file
	transfer.hash = sha256.New()
	session := sessionFor(transfer.deviceID)
	if session == nil {
		_ = file.Close()
		_ = os.Remove(temporary)
		return errors.New("ligação P2P indisponível")
	}
	payload, _ := json.Marshal(fileControl{Type: "file_accept", TransferID: transfer.id})
	if err := sendText(session, payload); err != nil {
		_ = file.Close()
		_ = os.Remove(temporary)
		return err
	}
	emitTransferEvent(eventFor(transfer, TransferStateReceiving, "a receber"))
	return nil
}

func RejectIncomingFile(transferID string) error {
	transfer := findTransfer(transferID)
	if transfer == nil || !transfer.incoming {
		return errors.New("transferência recebida não encontrada")
	}
	if session := sessionFor(transfer.deviceID); session != nil {
		payload, _ := json.Marshal(fileControl{Type: "file_reject", TransferID: transfer.id, Message: "ficheiro recusado"})
		_ = sendText(session, payload)
	}
	return cancelTransfer(transfer, TransferStateRejected, "ficheiro recusado", false)
}

func CancelTransfer(transferID string) error {
	transfer := findTransfer(transferID)
	if transfer == nil {
		return errors.New("transferência não encontrada")
	}
	return cancelTransfer(transfer, TransferStateCancelled, "transferência cancelada", true)
}

func cancelTransfer(transfer *fileTransfer, state TransferState, message string, notify bool) error {
	stopPendingTimeout(transfer)
	transfer.mu.Lock()
	if transfer.cancelled {
		transfer.mu.Unlock()
		return nil
	}
	transfer.cancelled = true
	file := transfer.file
	temporary := transfer.temporaryPath
	transfer.mu.Unlock()
	if file != nil {
		_ = file.Close()
	}
	if temporary != "" {
		_ = os.Remove(temporary)
	}
	if notify {
		if session := sessionFor(transfer.deviceID); session != nil {
			payload, _ := json.Marshal(fileControl{Type: "file_cancel", TransferID: transfer.id, Message: message})
			_ = sendText(session, payload)
		}
	}
	emitTransferEvent(eventFor(transfer, state, message))
	removeTransfer(transfer.id)
	return nil
}

func handleDataChannelMessage(session *peerSession, message webrtc.DataChannelMessage) {
	if message.IsString {
		handleFileControl(session, message.Data)
		return
	}
	handleFileChunk(session, message.Data)
}

func handleFileControl(session *peerSession, data []byte) {
	var control fileControl
	if err := json.Unmarshal(data, &control); err != nil {
		return
	}
	switch control.Type {
	case "file_offer":
		var offer fileOffer
		if json.Unmarshal(data, &offer) != nil || !validOffer(offer) {
			return
		}
		transfer := &fileTransfer{id: offer.TransferID, deviceID: session.targetDeviceID, name: filepath.Base(offer.Name), size: offer.Size, sha256: offer.SHA256, incoming: true}
		registerTransfer(transfer)
		setPendingTimeout(transfer)
		emitTransferEvent(eventFor(transfer, TransferStatePending, "pedido de receção"))
	case "file_accept":
		if transfer := findTransfer(control.TransferID); transfer != nil && !transfer.incoming {
			stopPendingTimeout(transfer)
			go sendFile(session, transfer)
		}
	case "file_reject":
		if transfer := findTransfer(control.TransferID); transfer != nil {
			_ = cancelTransfer(transfer, TransferStateRejected, "ficheiro recusado", false)
		}
	case "file_cancel", "file_error":
		if transfer := findTransfer(control.TransferID); transfer != nil {
			_ = cancelTransfer(transfer, TransferStateCancelled, control.Message, false)
		}
	case "file_complete":
		completeIncomingFile(session, control.TransferID)
	case "file_ack":
		if transfer := findTransfer(control.TransferID); transfer != nil && !transfer.incoming {
			stopPendingTimeout(transfer)
			emitTransferEvent(eventFor(transfer, TransferStateCompleted, "ficheiro entregue e confirmado"))
			removeTransfer(transfer.id)
		}
	}
}

func sendFile(session *peerSession, transfer *fileTransfer) {
	transfer.mu.Lock()
	if transfer.cancelled {
		transfer.mu.Unlock()
		return
	}
	transfer.mu.Unlock()
	file, err := os.Open(transfer.path)
	if err != nil {
		failTransfer(transfer, "não foi possível abrir o ficheiro")
		return
	}
	defer file.Close()
	emitTransferEvent(eventFor(transfer, TransferStateSending, "a enviar"))
	buffer := make([]byte, chunkSize)
	var sequence uint64
	for {
		read, readErr := file.Read(buffer)
		if read > 0 {
			if err := sendBinary(session, encodeChunk(transfer.id, sequence, buffer[:read])); err != nil {
				failTransfer(transfer, "erro ao enviar dados do ficheiro")
				return
			}
			sequence++
			transfer.mu.Lock()
			transfer.progress += int64(read)
			emitProgress(transfer, TransferStateSending)
			transfer.mu.Unlock()
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			failTransfer(transfer, "erro ao ler o ficheiro")
			return
		}
	}
	payload, _ := json.Marshal(fileControl{Type: "file_complete", TransferID: transfer.id})
	if err := sendText(session, payload); err != nil {
		failTransfer(transfer, "não foi possível terminar a transferência")
	}
}

func handleFileChunk(_ *peerSession, data []byte) {
	transferID, sequence, payload, err := decodeChunk(data)
	if err != nil {
		return
	}
	transfer := findTransfer(transferID)
	if transfer == nil || !transfer.incoming {
		return
	}
	stopPendingTimeout(transfer)
	transfer.mu.Lock()
	defer transfer.mu.Unlock()
	if transfer.file == nil || transfer.cancelled || sequence != transfer.nextChunk || int64(len(payload)) > transfer.size-transfer.progress {
		return
	}
	if _, err := transfer.file.Write(payload); err != nil {
		go failTransfer(transfer, "erro ao guardar o ficheiro")
		return
	}
	if _, err := transfer.hash.Write(payload); err != nil {
		go failTransfer(transfer, "erro ao validar o ficheiro")
		return
	}
	transfer.nextChunk++
	transfer.progress += int64(len(payload))
	emitProgress(transfer, TransferStateReceiving)
}

func completeIncomingFile(session *peerSession, transferID string) {
	transfer := findTransfer(transferID)
	if transfer == nil || !transfer.incoming {
		return
	}
	transfer.mu.Lock()
	if transfer.file == nil || transfer.progress != transfer.size || hex.EncodeToString(transfer.hash.Sum(nil)) != transfer.sha256 {
		transfer.mu.Unlock()
		failTransfer(transfer, "validação de integridade falhou")
		return
	}
	file := transfer.file
	temporary := transfer.temporaryPath
	destination := transfer.path
	transfer.file = nil
	transfer.mu.Unlock()
	if err := file.Close(); err != nil || os.Rename(temporary, destination) != nil {
		failTransfer(transfer, "não foi possível finalizar o ficheiro")
		return
	}
	payload, _ := json.Marshal(fileControl{Type: "file_ack", TransferID: transfer.id})
	if err := sendText(session, payload); err != nil {
		failTransfer(transfer, "ficheiro guardado, mas confirmação falhou")
		return
	}
	emitTransferEvent(eventFor(transfer, TransferStateCompleted, "ficheiro recebido"))
	removeTransfer(transfer.id)
}

func failTransfer(transfer *fileTransfer, message string) {
	_ = cancelTransfer(transfer, TransferStateFailed, message, true)
}

func cancelTransfersForDevice(deviceID, message string) {
	transferRegistry.RLock()
	transfers := make([]*fileTransfer, 0)
	for _, transfer := range transferRegistry.transfers {
		if transfer.deviceID == deviceID {
			transfers = append(transfers, transfer)
		}
	}
	transferRegistry.RUnlock()
	for _, transfer := range transfers {
		_ = cancelTransfer(transfer, TransferStateFailed, message, false)
	}
}

func setPendingTimeout(transfer *fileTransfer) {
	transfer.mu.Lock()
	if transfer.cancelled {
		transfer.mu.Unlock()
		return
	}
	if transfer.pendingTimer != nil {
		transfer.pendingTimer.Stop()
	}
	transfer.pendingTimer = time.AfterFunc(time.Minute, func() {
		_ = cancelTransfer(transfer, TransferStateFailed, "pedido de transferência expirou", true)
	})
	transfer.mu.Unlock()
}

func stopPendingTimeout(transfer *fileTransfer) {
	transfer.mu.Lock()
	if transfer.pendingTimer != nil {
		transfer.pendingTimer.Stop()
		transfer.pendingTimer = nil
	}
	transfer.mu.Unlock()
}

func validOffer(offer fileOffer) bool {
	if _, err := uuid.Parse(offer.TransferID); err != nil || offer.Size < 0 || offer.Size > maxTransferSize || filepath.Base(offer.Name) != offer.Name || strings.TrimSpace(offer.Name) == "" {
		return false
	}
	_, err := hex.DecodeString(offer.SHA256)
	return err == nil && len(offer.SHA256) == sha256.Size*2
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func encodeChunk(transferID string, sequence uint64, payload []byte) []byte {
	id := uuid.MustParse(transferID)
	data := make([]byte, chunkHeaderSize+len(payload))
	copy(data[:4], []byte{'S', 'A', 'V', '1'})
	data[4] = 1
	copy(data[5:21], id[:])
	binary.BigEndian.PutUint64(data[21:29], sequence)
	copy(data[29:], payload)
	return data
}

func decodeChunk(data []byte) (string, uint64, []byte, error) {
	if len(data) < chunkHeaderSize || string(data[:4]) != "SAV1" || data[4] != 1 {
		return "", 0, nil, errors.New("bloco inválido")
	}
	id, err := uuid.FromBytes(data[5:21])
	if err != nil {
		return "", 0, nil, err
	}
	return id.String(), binary.BigEndian.Uint64(data[21:29]), data[29:], nil
}

func availableFilePath(directory, name string) (string, error) {
	base := filepath.Base(name)
	if base != name || base == "." || base == string(filepath.Separator) {
		return "", errors.New("nome de ficheiro inválido")
	}
	extension := filepath.Ext(base)
	prefix := strings.TrimSuffix(base, extension)
	for number := 0; number < 10000; number++ {
		candidate := filepath.Join(directory, base)
		if number > 0 {
			candidate = filepath.Join(directory, fmt.Sprintf("%s (%d)%s", prefix, number, extension))
		}
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", errors.New("não foi possível obter um nome de ficheiro livre")
}

func eventFor(transfer *fileTransfer, state TransferState, message string) TransferEvent {
	return TransferEvent{ID: transfer.id, DeviceID: transfer.deviceID, Name: transfer.name, Size: transfer.size, Progress: transfer.progress, Incoming: transfer.incoming, State: state, Message: message}
}

func emitProgress(transfer *fileTransfer, state TransferState) {
	if time.Since(transfer.lastEvent) < 150*time.Millisecond && transfer.progress < transfer.size {
		return
	}
	transfer.lastEvent = time.Now()
	emitTransferEvent(eventFor(transfer, state, ""))
}
