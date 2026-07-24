package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowLogin(a fyne.App) {
	w := a.NewWindow("Login")
	w.Resize(fyne.NewSize(400, 250))

	email := widget.NewEntry()
	email.SetPlaceHolder("Enter your email")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Enter your password")

	loginButton := widget.NewButton("Login", func() {
		fmt.Println("Email:", email.Text)
		fmt.Println("Password:", password.Text)

		if helpers.Login(email.Text, password.Text) {
			fmt.Println("Login successful!")
			ShowHome(a)
		} else {
			fmt.Println("Login failed!")
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Login"),
		email,
		password,
		loginButton,
	)

	w.SetContent(form)
	w.Show()
}
