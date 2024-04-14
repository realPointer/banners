package v1

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/realPointer/banners/internal/controller/http/v1/middlewares"
	"github.com/realPointer/banners/internal/service"
	"github.com/rs/zerolog"
)

func NewRouter(l *zerolog.Logger, Services *service.Services) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(time.Minute))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/auth", NewAuthRouter(Services.Auth, l))

		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(Services.Auth))
			r.Mount("/", NewBannerRouter(Services.Banner, l))
		})
	})

	return router
}

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

			ctx := context.WithValue(r.Context(), "role", claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
