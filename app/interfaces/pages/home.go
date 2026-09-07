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

	content := container.NewVBox(
		widget.NewLabel("Home Page"),
		usersbouton,
		usersLabel,
	)

	w.SetContent(content)
	w.Show()
}
