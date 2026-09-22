package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	content := container.NewVBox(gobackButton, userLabel)
	w.SetContent(content)
	w.Show()
}