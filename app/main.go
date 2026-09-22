package main

import (
	"app/interfaces/pages"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"app/core/helpers"
)

func main() {
	a := app.New()
	fmt.Println("App module is running")

	w := a.NewWindow("SaveApp")
	w.Resize(fyne.NewSize(500, 400))

	pages.ShowLogin(w, a)

	w.ShowAndRun()
	helpers.StopWebSocket()
}
