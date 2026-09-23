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

	var devices2 []fyne.CanvasObject

	for _, device := range devices {
		deviceID := fmt.Sprintf("%v", device["id"])

		deviceLabel := widget.NewButton(
			fmt.Sprintf(
				"ID: %v\n Nome: %v\n LastSeenAt: %v\n",
				device["id"],
				device["name"],
				device["last_seen_at"],
			),
			func() {
				err := helpers.SendTestMessage(
					deviceID,
					"Olá! Mensagem de teste.",
				)

				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					return
				}

				fmt.Println("Mensagem de teste enviada!")
			},
		)

		devices2 = append(devices2, deviceLabel)
	}

	if len(devices2) == 0 {
		devices2 = append(devices2, widget.NewLabel("Nenhum device encontrado"))
	}

	userInfo := fmt.Sprintf(
		"ID: %v\nNome: %v\nEmail: %v",
		user["id"],
		user["name"],
		user["email"],
	)

	userLabel := widget.NewLabel(userInfo)

	gobackButton := widget.NewButton("Go back", func() {
		ShowSearchUsers(w, a)
	})

	content := container.NewVBox(
		gobackButton,
		userLabel,
		container.New(
			layout.NewAdaptiveGridLayout(5),
			devices2...,
		),
	)

	w.SetContent(content)
	w.Show()
}
