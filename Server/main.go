package main

import (
	"fmt"
	"net/http"

	"Server/config"
	"Server/controllers"
	"Server/repositories"
	"Server/routes"
	"Server/services"
)

func main() {
	db, err := config.ConnectDatabase()
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)

	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	authService := services.NewAuthService(userRepo)
	authController := controllers.NewAuthController(authService)

	deviceService := services.NewDeviceService(userRepo)
	deviceController := controllers.NewDeviceController(deviceService)
	webSocketController := controllers.NewWebSocketController(deviceService, controllers.NewConnectionHub())

	routes.RegisterRoutes(userController, authController, deviceController, webSocketController)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
