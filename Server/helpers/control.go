package helpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetTokenFromCookie(r)
		if err != nil {
			JSON(w, http.StatusUnauthorized, "Not logged in")
			return
		}

		userID, err := ValidateJWT(token)
		if err != nil {
			JSON(w, http.StatusUnauthorized, "Not logged in")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthenticatedUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func HasValidSession(r *http.Request) bool {
	token, err := GetTokenFromCookie(r)
	if err != nil {
		return false
	}

	_, err = ValidateJWT(token)
	return err == nil
}

func GenerateDeviceID() (string, error) {
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	), nil
}
