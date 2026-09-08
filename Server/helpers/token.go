package helpers

import (
	"errors"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secretKey = []byte(jwtSecret())

func GenerateJWT(userID int64) (string, error) {

	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   cookieSecure(),
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func ValidateJWT(tokenString string) (int64, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	userIDValue, ok := claims["userID"].(float64)
	if !ok || userIDValue <= 0 || userIDValue != math.Trunc(userIDValue) {
		return 0, errors.New("invalid user id claim")
	}

	return int64(userIDValue), nil
}

func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   cookieSecure(),
		Expires:  time.Unix(0, 0),
	})
}

func cookieSecure() bool {
	loadDotEnv()
	secure, _ := strconv.ParseBool(os.Getenv("COOKIE_SECURE"))
	return secure
}
