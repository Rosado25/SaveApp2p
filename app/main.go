package main

import (
	"app/interfaces/pages"
	"fmt"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	fmt.Println("App module is running")
	pages.ShowLogin(a)
	a.Run()
}
