package routes

import (
	"Server/controllers"
	"net/http"
)

func AuthRoutes(controller *controllers.AuthController) {
	http.HandleFunc("/login", controller.Login)
	http.HandleFunc("/logout", controller.Logout)
	http.HandleFunc("/token", controller.Token)
	http.HandleFunc("/CheckToken", controller.CheckConnection)
	//http.HandleFunc("/register", controller.Register)
}
