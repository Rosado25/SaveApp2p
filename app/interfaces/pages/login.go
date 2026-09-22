package pages

import (
	"app/core/helpers"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowLogin(w fyne.Window, a fyne.App) {
	w.SetTitle("Login")

	email := widget.NewEntry()
	email.SetPlaceHolder("Enter your email")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Enter your password")

	loginButton := widget.NewButton("Login", func() {
		fmt.Println("Email:", email.Text)
		fmt.Println("Password:", password.Text)

		loggedIn, err := helpers.Login(email.Text, password.Text)
		if loggedIn {
			fmt.Println("Login successful!")
			if errors.Is(err, helpers.ErrDeviceSetupRequired) {
				ShowCreateDevice(w, a)
				return
			}
			ShowHome(w, a)
		} else {
			fmt.Println("Login failed:", err)
		}
	})

	registerButton := widget.NewButton("Create account", func() {
		ShowRegister(w, a)
	})

	form := container.NewVBox(
		widget.NewLabel("Login"),
		email,
		password,
		loginButton,
		registerButton,
	)

	w.SetContent(form)
	w.Show()
}
