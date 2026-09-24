package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	deviceIDs := make([]string, 0, len(devices))
	deviceNames := make([]string, 0, len(devices))

	for _, device := range devices {
		deviceID := fmt.Sprintf("%v", device["id"])
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

	messageInput := container.NewBorder(
		nil,
		nil,
		nil,
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
		deviceLabel := widget.NewLabel(
			fmt.Sprintf(
				"ID: %v\nNome: %v\nLastSeenAt: %v",
				device["id"],
				device["name"],
				device["last_seen_at"],
			),
		)

		devices2 = append(devices2, deviceLabel)
	}

	if len(devices2) == 0 {
		devices2 = append(
			devices2,
			widget.NewLabel("Nenhum device encontrado"),
		)
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
	)

	w.SetContent(content)
	w.Show()
}
