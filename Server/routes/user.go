package routes

import (
	"Server/controllers"
	"net/http"
)

func UserRoutes(controller *controllers.UserController) {
	http.HandleFunc("/CreateUser", controller.CreateUser)
	http.HandleFunc("/GetUsers", controller.GetUsers)
	http.HandleFunc("/GetUserByID", controller.GetUserByID)
	http.HandleFunc("/GetUserByEmail", controller.GetUserByEmail)
}
