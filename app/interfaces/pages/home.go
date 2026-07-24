package pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowHome(a fyne.App) {
	w := a.NewWindow("Home")
	w.Resize(fyne.NewSize(400, 250))

	content := container.NewVBox(
		widget.NewLabel("Home Page"),
	)

	w.SetContent(content)
	w.Show()
}
