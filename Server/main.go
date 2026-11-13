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

	authService := services.NewAuthService(userRepo)
	authController := controllers.NewAuthController(authService)

	connectionHub := controllers.NewConnectionHub()

	deviceService := services.NewDeviceService(userRepo)
	deviceController := controllers.NewDeviceController(deviceService, connectionHub)

	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService, deviceService, connectionHub)

	webSocketController := controllers.NewWebSocketController(
		deviceService,
		connectionHub,
	)

	routes.RegisterRoutes(userController, authController, deviceController, webSocketController)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
