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

func GetDeviceByID(deviceID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiURL("/GetDeviceByID?id="+deviceID), nil)
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

	var device map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func GetDevicesFromUser(userID string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", apiURL("/DevicesFromUser?user_id="+userID), nil)
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

func DeviceIdToName(deviceID string) (string, error) {

	req, err := http.NewRequest("GET", apiURL("/DeviceIdToName?device_id="+deviceID), nil)
	if err != nil {
		return "", err
	}

	token := GetToken()
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}

	resp, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	var name string

	if err := json.NewDecoder(resp.Body).Decode(&name); err != nil {
		return "", err
	}

	return name, nil
}
