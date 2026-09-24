package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowDevice(w fyne.Window, a fyne.App, deviceID string) {
	w.SetTitle("Device Profile")

	device, err := helpers.GetDeviceByID(deviceID)
	if err != nil {
		fmt.Println("Erro ao carregar device:", err)
		w.SetContent(widget.NewLabel("Erro ao carregar device: " + err.Error()))
		return
	}

	devicInfo := fmt.Sprintf(
		"ID: %v\nNome: %v\nDevice ID: %v\n LastSeenAt: %v\n CreatedAt: %v\n",
		device["id"],
		device["name"],
		device["device_id"],
		device["last_seen_at"],
		device["created_at"],
	)

	deviceLabel := widget.NewLabel(devicInfo)

	testButton := widget.NewButton("Test WebSocket", func() {
		err := helpers.SendMessage(deviceID, "Olá! Mensagem de teste.")
		if err != nil {
			fmt.Println("Erro ao enviar mensagem:", err)
			return
		}

		fmt.Println("Mensagem de teste enviada!")
	})

	gobackButton := widget.NewButton("Go back", func() {
		ShowHome(w, a)
	})

	content := container.NewVBox(gobackButton, deviceLabel, testButton)
	w.SetContent(content)
	w.Show()
}
