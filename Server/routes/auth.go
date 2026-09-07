package routes

import (
	"Server/controllers"
	"Server/helpers"
	"net/http"
)

func AuthRoutes(controller *controllers.AuthController) {
	http.HandleFunc("/login", controller.Login)
	http.Handle("/logout", helpers.RequireAuth(http.HandlerFunc(controller.Logout)))
	http.Handle("/token", helpers.RequireAuth(http.HandlerFunc(controller.Token)))
	http.Handle("/CheckToken", helpers.RequireAuth(http.HandlerFunc(controller.CheckConnection)))
	//http.HandleFunc("/register", controller.Register)
}
