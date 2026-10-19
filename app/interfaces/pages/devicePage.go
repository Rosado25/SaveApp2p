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

	targetDeviceID := fmt.Sprintf("%v", device["device_id"])
	deviceLabel := widget.NewButton(devicInfo, func() {
		_, err := helpers.StartWebRTCSender(targetDeviceID)

		if err != nil {
			fmt.Println("Erro ao iniciar WebRTC:", err)
			return
		}

		fmt.Println("WebRTC iniciado!")
	})

	gobackButton := widget.NewButton("Go back", func() {
		ShowHome(w, a)
	})

	content := container.NewVBox(gobackButton, deviceLabel)
	w.SetContent(content)
	w.Show()
}
