package helpers

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func StringToInt64(s string) (int64, error) {
	var i int64
	err := json.Unmarshal([]byte(s), &i)
	return i, err
}
