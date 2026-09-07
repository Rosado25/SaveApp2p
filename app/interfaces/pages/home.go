package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowHome(a fyne.App) {
	w := a.NewWindow("Home")
	w.Resize(fyne.NewSize(400, 250))

	usersLabel := widget.NewLabel("Clica em GetUsers para carregar os utilizadores")

	usersbouton := widget.NewButton("GetUsers", func() {
		users, err := helpers.GetUsers()
		if err != nil {
			usersLabel.SetText("Erro ao carregar utilizadores: " + err.Error())
			return
		}

		text := ""
		for _, user := range users {
			text += fmt.Sprintf("ID: %v | Nome: %v | Email: %v\n", user["id"], user["name"], user["email"])
		}
		if text == "" {
			text = "Nenhum utilizador encontrado"
		}
		usersLabel.SetText(text)
	})

	devicesLabel := widget.NewLabel("A carregar devices...")
	loadDevices := func() {
		devices, err := helpers.GetDevicesByUser()
		if err != nil {
			devicesLabel.SetText("Erro ao carregar devices: " + err.Error())
			return
		}

		text := ""
		for _, device := range devices {
			text += fmt.Sprintf("ID: %v | Nome: %v | DeviceID: %v\n", device["id"], device["name"], device["device_id"])
		}
		if text == "" {
			text = "Nenhum device encontrado"
		}
		devicesLabel.SetText(text)
	}

	loadDevices()

	logoutbouton := widget.NewButton("Logout", func() {
		if helpers.Logout() {
			fmt.Println("Logout successful!")
			w.Close()
			ShowLogin(a)
		} else {
			fmt.Println("Erro ao fazer logout")
		}
	})

	closebouton := widget.NewButton("Close", func() {
		if helpers.Logout() {
			fmt.Println("Bye Bye!")
			w.Close()
		} else {
			fmt.Println("Erro a fechar a aplicação")
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Home Page"),
		closebouton,
		logoutbouton,
		usersbouton,
		usersLabel,
		widget.NewLabel("Devices do utilizador logado:"),
		devicesLabel,
	)

	w.SetContent(content)
	w.Show()
}
