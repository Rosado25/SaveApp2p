package routes

import "Server/controllers"

func RegisterRoutes(userController *controllers.UserController, authController *controllers.AuthController, deviceController *controllers.DeviceController) {
	UserRoutes(userController)
	AuthRoutes(authController)
	DeviceRoutes(deviceController)
}
