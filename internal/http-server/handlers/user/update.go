package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

type UpdateUserRequest struct {
	Name string `json:"name"`
}

func (i *Implementation) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный ID пользователя"}})
		return
	}
	var request UpdateUserRequest
	if err := handlers.DecodeJSON(r, &request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный JSON"}})
		return
	}
	result, err := i.userService.UpdateUser(r.Context(), id, request.Name)
	if err != nil {
		if errors.Is(err, service.ErrInvalidName) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: err.Error()}})
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorNotFound, Message: err.Error()}})
			return
		}
		logger.Error("failed to update user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
