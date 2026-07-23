package routes

import (
	"Server/controllers"
	"net/http"
)

func AuthRoutes(controller *controllers.AuthController) {
	http.HandleFunc("/login", controller.Login)
	http.HandleFunc("/logout", controller.Logout)
	//http.HandleFunc("/register", controller.Register)
}
