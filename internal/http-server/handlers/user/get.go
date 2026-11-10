package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

func (i *Implementation) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный ID пользователя"}})
		return
	}
	result, err := i.userService.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorNotFound, Message: err.Error()}})
			return
		}
		logger.Error("failed to get user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
