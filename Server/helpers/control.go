package helpers

import (
	"net/http"
)

func CheckConnection(w http.ResponseWriter, r *http.Request, t int, message string) bool {
	token, _ := r.Cookie("token")
	switch t {
	case 0:
		if token != nil {
			if _, err := ValidateJWT(token.Value); err != nil {
				return false
			}
			JSON(w, 200, message)
			return true
		}
	case 1:
		if token == nil {
			JSON(w, 401, message)
			return true
		}
		if _, err := ValidateJWT(token.Value); err != nil {
			JSON(w, 401, message)
			return true
		}
	default:
		return false
	}

	return false
}
