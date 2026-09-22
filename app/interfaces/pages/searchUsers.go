package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowSearchUsers(w fyne.Window ,a fyne.App) {

	w.SetTitle("Search Users")

	users, err := helpers.GetUsers()
	if err != nil {
		fmt.Println("Erro ao carregar utilizadores:", err)
	}

	var userItems []fyne.CanvasObject

	for _, user := range users {
		userLabel := widget.NewButton(
			fmt.Sprintf(
				"ID: %v\nNome: %v\nEmail: %v",
				user["id"],
				user["name"],
				user["email"],
			), func() {
				
				id := fmt.Sprintf("%v", user["id"])
				ShowUserProfile(w, a, id)
			},
		)

		userItems = append(userItems, userLabel)
	}

	if len(userItems) == 0 {
		userItems = append(userItems, widget.NewLabel("Nenhum utilizador encontrado"))
	}

	grid := container.New(
		layout.NewAdaptiveGridLayout(5),
		userItems...,
	)

	gobackButton := widget.NewButton("Go back", func() {
		ShowHome(w,a)
	})

	content := container.NewVBox(gobackButton, grid)

	w.SetContent(content)
	w.Show()
}
