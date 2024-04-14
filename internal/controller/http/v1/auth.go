package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/realPointer/banners/internal/service"
	"github.com/rs/zerolog"
)

type authRoutes struct {
	authService service.Auth
	l           *zerolog.Logger
}

func NewAuthRouter(authService service.Auth, l *zerolog.Logger) http.Handler {
	s := &authRoutes{
		authService: authService,
		l:           l,
	}
	r := chi.NewRouter()

	r.Post("/register", s.registerHandler)
	r.Post("/login", s.loginHandler)

	return r
}

func (s *authRoutes) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	if err := s.authService.Register(r.Context(), req.Username, req.Password, req.Role); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *authRoutes) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token, err := s.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
