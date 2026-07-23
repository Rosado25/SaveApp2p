package routes

import "Server/controllers"

func RegisterRoutes(userController *controllers.UserController, authController *controllers.AuthController) {
	UserRoutes(userController)
	AuthRoutes(authController)
}
