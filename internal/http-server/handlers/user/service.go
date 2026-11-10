package user

import (
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type Implementation struct {
	userService service.UserService
}

func NewImplementation(s service.UserService) *Implementation {
	return &Implementation{userService: s}
}

func (i *Implementation) RegisterRoutes(mux *http.ServeMux, auth *middleware.Auth) {
	mux.HandleFunc("POST /api/v1/users", i.CreateUser)
	mux.HandleFunc("GET /api/v1/users", auth.RequireAdmin(i.GetUsers))
	mux.HandleFunc("GET /api/v1/users/{user_id}", auth.RequireOwner(i.GetUser))
	mux.HandleFunc("PATCH /api/v1/users/{user_id}", auth.RequireOwner(i.UpdateUser))
	mux.HandleFunc("DELETE /api/v1/users/{user_id}", auth.RequireOwner(i.DeleteUser))
}
