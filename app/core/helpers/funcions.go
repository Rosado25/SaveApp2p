package helpers

import (
	"encoding/json"
	"net/http"
)

func GetUsers() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiURL("/GetUsers"), nil)
	if err != nil {
		return nil, err
	}

	token := GetToken()
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func GetDevicesByUser() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiURL("/GetDevicesByUser"), nil)
	if err != nil {
		return nil, err
	}

	token := GetToken()
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var devices []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}


func GetUserByID(userID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiURL("/GetUserByID?id="+userID), nil)
	if err != nil {
		return nil, err
	}

	token := GetToken()
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}