package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowHome(w fyne.Window, a fyne.App) {

	if err := helpers.StartWebSocket(); err != nil {
		fmt.Println("WebSocket could not be started:", err)
	}

	w.SetTitle("Home")

	currentDeviceLabel := widget.NewLabel("Device em login: a carregar...")
	var devices2 []fyne.CanvasObject
	loadDevices := func() {
		devices, err := helpers.GetDevicesByUser()
		if err != nil {
			currentDeviceLabel.SetText("Device em login: indisponível")
			return
		}

		currentDeviceID, err := helpers.GetDeviceID()
		if err != nil || currentDeviceID == "" {
			currentDeviceLabel.SetText("Device em login: desconhecido")
		}

		for _, device := range devices {
			if device["device_id"] == currentDeviceID {
				currentDeviceLabel.SetText(fmt.Sprintf("Device em login: %v ", device["name"]))
			}
			deviceLabel := widget.NewButton(
				fmt.Sprintf(
					"ID: %v\n Nome: %v\n",
					device["id"],
					device["name"],
				), func() {
					id := fmt.Sprintf("%v", device["id"])
					ShowDevice(w, a, id)
				},
			)

			devices2 = append(devices2, deviceLabel)
		}

		if len(devices2) == 0 {
			devices2 = append(devices2, widget.NewLabel("Nenhum device encontrado"))
		}
	}

	loadDevices()

	logoutbouton := widget.NewButton("Logout", func() {
		if helpers.Logout() {
			fmt.Println("Logout successful!")
			ShowLogin(w, a)
		} else {
			fmt.Println("Erro ao fazer logout")
		}
	})

	SearchUsers := widget.NewButton("Search Users", func() {
		ShowSearchUsers(w, a)
	})

	SearchUsers.Resize(fyne.NewSize(100, 50))

	middle := container.NewWithoutLayout(SearchUsers)

	top := container.NewHBox(
		widget.NewLabel("Home Page"),
		layout.NewSpacer(),
		currentDeviceLabel,
		layout.NewSpacer(),
		logoutbouton,
	)

	bottom := container.NewVBox(
		widget.NewLabel("Devices do utilizador logado:"),
		container.New(
			layout.NewAdaptiveGridLayout(5),
			devices2...,
		),
	)

	content := container.New(
		layout.NewBorderLayout(
			top,
			bottom,
			nil,
			nil,
		),
		top,
		bottom,
		middle,
	)

	w.SetContent(content)
	w.Show()
}
