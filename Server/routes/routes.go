package routes

import (
	"Server/controllers"
	"net/http"
)

func RegisterRoutes(userController *controllers.UserController, authController *controllers.AuthController, deviceController *controllers.DeviceController, webSocketController *controllers.WebSocketController) {
	UserRoutes(userController)
	AuthRoutes(authController)
	DeviceRoutes(deviceController)
	http.HandleFunc("/ws", webSocketController.Handle)
}
