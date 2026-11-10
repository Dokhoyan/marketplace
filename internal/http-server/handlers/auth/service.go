package auth

import (
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type Implementation struct {
	authService service.AuthService
}

func NewImplementation(s service.AuthService) *Implementation {
	return &Implementation{authService: s}
}

func (i *Implementation) RegisterRoutes(mux *http.ServeMux, auth *middleware.Auth) {
	mux.HandleFunc("POST /api/v1/auth/login", i.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", auth.RequireAuth(i.Logout))
}
