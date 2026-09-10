package pages

import (
	"app/core/helpers"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowCreateDevice(a fyne.App) {
	w := a.NewWindow("Create Device")
	w.Resize(fyne.NewSize(400, 300))

	name := widget.NewEntry()
	name.SetPlaceHolder("Device name, e.g. Office PC")
	status := widget.NewLabel("Choose a name for this computer.")

	createButton := widget.NewButton("Create device", func() {
		if name.Text == "" {
			status.SetText("Device name is required.")
			return
		}

		if _, err := helpers.CreateDevice(name.Text); err != nil {
			status.SetText("Could not create device: " + err.Error())
			return
		}

		fmt.Println("Device created:", name.Text)
		if err := helpers.StartWebSocket(); err != nil {
			fmt.Println("WebSocket could not be started:", err)
		}
		w.Close()
		ShowHome(a)
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Create your device"),
		name,
		createButton,
		status,
	))
	w.Show()
}
