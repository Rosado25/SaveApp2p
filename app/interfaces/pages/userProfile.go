package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowUserProfile(w fyne.Window, a fyne.App, userID string) {
	w.SetTitle("User Profile")

	user, err := helpers.GetUserByID(userID)
	if err != nil {
		fmt.Println("Erro ao carregar utilizador:", err)
		w.SetContent(widget.NewLabel("Erro ao carregar utilizador: " + err.Error()))
		return
	}

	devices, err := helpers.GetDevicesFromUser(userID)
	if err != nil {
		fmt.Println("Erro ao carregar devices do utilizador:", err)
		w.SetContent(widget.NewLabel("Erro ao carregar devices do utilizador: " + err.Error()))
		return
	}

	// -------------------------
	// USER INFO
	// -------------------------

	userInfo := fmt.Sprintf(
		"ID: %v\nNome: %v\nEmail: %v",
		user["id"],
		user["name"],
		user["email"],
	)

	userLabel := widget.NewLabel(userInfo)

	// -------------------------
	// DEVICE SELECTION
	// -------------------------

	messageDisplay := widget.NewMultiLineEntry()
	messageDisplay.SetPlaceHolder("As mensagens aparecerão aqui...")
	messageDisplay.Disable()

	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Digite sua mensagem aqui...")
	transferStatus := widget.NewLabel("Transferências: nenhuma ativa")
	activeTransferID := ""

	deviceIDs := make([]string, 0, len(devices))
	deviceNames := make([]string, 0, len(devices))

	for _, device := range devices {
		deviceID := fmt.Sprintf("%v", device["device_id"])
		deviceName := fmt.Sprintf("%v", device["name"])

		deviceIDs = append(deviceIDs, deviceID)
		deviceNames = append(
			deviceNames,
			fmt.Sprintf(
				"%s (ID: %s)",
				deviceName,
				deviceID,
			),
		)
	}

	selectedDeviceID := ""

	deviceSelect := widget.NewSelect(
		deviceNames,
		func(selected string) {
			for i, name := range deviceNames {
				if name == selected {
					selectedDeviceID = deviceIDs[i]

					fmt.Println(
						"Device selecionado:",
						selectedDeviceID,
					)

					break
				}
			}
		},
	)

	deviceSelect.PlaceHolder = "Selecione um device"

	// -------------------------
	// MESSAGES
	// -------------------------

	helpers.SetChatMessageHandler(func(message helpers.WebSocketMessage) {

		senderName, err := helpers.DeviceIdToName(message.SenderDeviceID)
		if err != nil {
			fmt.Println("Erro ao obter nome do device:", err)
			senderName = "Device desconhecido"
		}

		receivedMessage := fmt.Sprintf("%v", message.Data)

		fyne.Do(func() {
			newMessage := senderName + ": " + receivedMessage

			if messageDisplay.Text == "" {
				messageDisplay.SetText(newMessage)
			} else {
				messageDisplay.SetText(
					messageDisplay.Text + "\n" + newMessage,
				)
			}
		})
	})

	helpers.SetTransferEventHandler(func(event helpers.TransferEvent) {
		fyne.Do(func() {
			progress := ""
			if event.Size > 0 {
				progress = fmt.Sprintf(" %.1f%%", float64(event.Progress)*100/float64(event.Size))
			}
			transferStatus.SetText(fmt.Sprintf("%s: %s%s — %s", event.Name, event.State, progress, event.Message))

			if event.State == helpers.TransferStatePending && event.Incoming {
				dialog.ShowConfirm(
					"Receber ficheiro",
					fmt.Sprintf("%s (%d bytes). Pretende receber este ficheiro?", event.Name, event.Size),
					func(accepted bool) {
						if !accepted {
							_ = helpers.RejectIncomingFile(event.ID)
							return
						}
						dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
							if err != nil || folder == nil {
								_ = helpers.RejectIncomingFile(event.ID)
								return
							}
							if err := helpers.AcceptIncomingFile(event.ID, folder.Path()); err != nil {
								transferStatus.SetText("Erro ao preparar a receção: " + err.Error())
								return
							}
							activeTransferID = event.ID
						}, w)
					},
					w,
				)
			}
		})
	})

	sendButton := widget.NewButton("Enviar", func() {
		if selectedDeviceID == "" {
			fmt.Println("Nenhum device selecionado")
			return
		}

		message := messageEntry.Text

		if message == "" {
			return
		}

		err := helpers.SendMessage(
			selectedDeviceID,
			message,
		)

		if err != nil {
			fmt.Println("Erro ao enviar mensagem:", err)
			return
		}

		if messageDisplay.Text == "" {
			messageDisplay.SetText(
				"Você: " + message,
			)
		} else {
			messageDisplay.SetText(
				messageDisplay.Text + "\nVocê: " + message,
			)
		}

		messageEntry.SetText("")

		fmt.Println(
			"Mensagem enviada para device:",
			selectedDeviceID,
		)
	})

	fileButton := widget.NewButton("Enviar ficheiro", func() {
		if selectedDeviceID == "" {
			transferStatus.SetText("Selecione um device antes de enviar um ficheiro")
			return
		}
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			path := reader.URI().Path()
			_ = reader.Close()
			transferID, err := helpers.SendFile(selectedDeviceID, path)
			if err != nil {
				transferStatus.SetText("Erro ao iniciar transferência: " + err.Error())
				return
			}
			activeTransferID = transferID
		}, w)
	})

	cancelTransferButton := widget.NewButton("Cancelar transferência", func() {
		if activeTransferID == "" {
			return
		}
		if err := helpers.CancelTransfer(activeTransferID); err != nil {
			transferStatus.SetText("Erro ao cancelar: " + err.Error())
		}
	})

	messageInput := container.NewBorder(
		nil,
		nil,
		fileButton,
		sendButton,
		messageEntry,
	)

	chat := container.NewBorder(
		nil,
		messageInput,
		nil,
		nil,
		messageDisplay,
	)

	// -------------------------
	// DEVICES INFO
	// -------------------------

	var devices2 []fyne.CanvasObject

	for _, device := range devices {
		deviceID := fmt.Sprintf("%v", device["device_id"])
		targetDeviceID := deviceID

		deviceLabel := widget.NewButton(
			fmt.Sprintf(
				"ID: %v\nNome: %v\nLastSeenAt: %v",
				device["id"],
				device["name"],
				device["last_seen_at"],
			),
			func() {
				_, err := helpers.StartWebRTCSender(targetDeviceID)

				if err != nil {
					fmt.Println("Erro ao iniciar WebRTC:", err)
					return
				}

				fmt.Println("WebRTC iniciado!")
			},
		)

		devices2 = append(devices2, deviceLabel)
	}

	// -------------------------
	// GO BACK
	// -------------------------

	gobackButton := widget.NewButton("Go back", func() {
		ShowSearchUsers(w, a)
	})

	// -------------------------
	// CONTENT
	// -------------------------

	content := container.NewVBox(
		gobackButton,

		userLabel,

		widget.NewSeparator(),

		widget.NewLabel("Devices:"),
		container.New(
			layout.NewAdaptiveGridLayout(5),
			devices2...,
		),

		widget.NewSeparator(),

		widget.NewLabel("Enviar mensagem para:"),
		deviceSelect,

		widget.NewLabel("Mensagens:"),
		chat,

		widget.NewSeparator(),
		transferStatus,
		cancelTransferButton,
	)

	w.SetContent(content)
	w.Show()
}
