package routes

import (
	"Server/controllers"
	"net/http"
)

func DeviceRoutes(controller *controllers.DeviceController) {
	http.HandleFunc("/CreateDevice", controller.CreateDevice)
}
