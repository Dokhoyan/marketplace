package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (i *Implementation) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest
	if err := handlers.DecodeJSON(r, &request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный JSON"},
		})
		return
	}

	created, err := i.userService.CreateUser(r.Context(), request.Name, request.Login, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidName) ||
			errors.Is(err, service.ErrInvalidLogin) ||
			errors.Is(err, service.ErrInvalidPassword) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: err.Error()},
			})
			return
		}

		if errors.Is(err, service.ErrLoginTaken) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error: model.ErrorBody{Code: model.ErrorLoginTaken, Message: err.Error()},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		logger.Error("failed to create user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}
