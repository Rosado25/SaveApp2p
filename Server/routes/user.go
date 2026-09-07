package routes

import (
	"Server/controllers"
	"Server/helpers"
	"net/http"
)

func UserRoutes(controller *controllers.UserController) {
	http.Handle("/CreateUser", helpers.RequireAuth(http.HandlerFunc(controller.CreateUser)))
	http.Handle("/GetUsers", helpers.RequireAuth(http.HandlerFunc(controller.GetUsers)))
	http.Handle("/GetUserByID", helpers.RequireAuth(http.HandlerFunc(controller.GetUserByID)))
	http.Handle("/GetUserByEmail", helpers.RequireAuth(http.HandlerFunc(controller.GetUserByEmail)))
}
