package main

import (
	"app/interfaces/pages"
	"fmt"

	"app/core/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	fmt.Println("App module is running")

	w := a.NewWindow("SaveApp")
	w.Resize(fyne.NewSize(750, 600))

	pages.ShowLogin(w, a)

	w.ShowAndRun()
	helpers.StopWebSocket()
}
