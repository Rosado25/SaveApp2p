package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowRegister(a fyne.App) {
	w := a.NewWindow("Register")
	w.Resize(fyne.NewSize(400, 300))

	email := widget.NewEntry()
	email.SetPlaceHolder("Enter your email")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Enter your password")
	status := widget.NewLabel("")

	registerButton := widget.NewButton("Register", func() {
		fmt.Println("Email:", email.Text)
		fmt.Println("Password:", password.Text)

		if helpers.Register(email.Text, password.Text) {
			fmt.Println("Register successful!")
			w.Close()
			ShowLogin(a)
		} else {
			fmt.Println("Register failed!")
			status.SetText("Registration failed. Check the data or email.")
		}
	})

	gobackButton := widget.NewButton("Go back", func() {
		w.Close()
		ShowLogin(a)
	})

	form := container.NewVBox(
		widget.NewLabel("Register"),
		email,
		password,
		registerButton,
		gobackButton,
		status,
	)

	w.SetContent(form)
	w.Show()
}
