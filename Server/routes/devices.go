package routes

import (
	"Server/controllers"
	"Server/helpers"
	"net/http"
)

func DeviceRoutes(controller *controllers.DeviceController) {
	http.Handle("/CreateDevice", helpers.RequireAuth(http.HandlerFunc(controller.CreateDevice)))
	http.Handle("/GetDevicesByUser", helpers.RequireAuth(http.HandlerFunc(controller.GetDevicesByUser)))
	http.Handle("/GetDeviceByID", helpers.RequireAuth(http.HandlerFunc(controller.GetDeviceByID)))
	http.Handle("/DevicesFromUser", helpers.RequireAuth(http.HandlerFunc(controller.DevicesFromUser)))
}
