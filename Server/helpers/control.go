package helpers

import (
	"context"
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
