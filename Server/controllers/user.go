package controllers

import (
	"Server/helpers"
	"Server/models"
	"Server/services"
	"encoding/json"
	"net/http"
)

type UserController struct {
	service *services.UserService
}

func NewUserController(s *services.UserService) *UserController {
	return &UserController{service: s}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		helpers.JSON(w, 400, "Invalid request payload")
		return
	}

	if err := c.service.CreateUser(user); err != nil {
		helpers.JSON(w, 500, err.Error())
		return
	}

	// Never return the password, even though the service stores its hash.
	user.Password = ""
	helpers.JSON(w, 201, user)
}

func (c *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	response, err := c.service.GetUsers()
	if err != nil {
		helpers.JSON(w, 500, err.Error())
		return
	}
	for i := range response {
		response[i].Password = ""
	}

	helpers.JSON(w, 200, response)
}

func (c *UserController) GetUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		helpers.JSON(w, 400, "Missing user ID")
		return
	}

	response, err := c.service.GetUserByID(id)
	if err != nil {
		helpers.JSON(w, 500, err.Error())
		return
	}
	response.Password = ""

	helpers.JSON(w, 200, response)
}

func (c *UserController) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.JSON(w, 405, "Method not allowed")
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		helpers.JSON(w, 400, "Missing user email")
		return
	}

	response, err := c.service.GetUserByEmail(email)
	if err != nil {
		helpers.JSON(w, 500, err.Error())
		return
	}
	response.Password = ""

	helpers.JSON(w, 200, response)
}
