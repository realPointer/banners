package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/realPointer/banners/internal/service"
)

type contextKey string

const roleContextKey contextKey = "role"

func AuthMiddleware(authService service.Auth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

			claims, err := authService.ParseToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), roleContextKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(roleContextKey).(string)
		if role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
