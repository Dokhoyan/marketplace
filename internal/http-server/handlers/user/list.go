package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"go.uber.org/zap"
)

type UserSummary struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Login     string    `json:"login"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (i *Implementation) GetUsers(w http.ResponseWriter, r *http.Request) {
	result, err := i.userService.GetUsers(r.Context())
	if err != nil {
		logger.Error("failed to list user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"}})
		return
	}
	response := make([]UserSummary, 0, len(result))
	for _, u := range result {
		response = append(response, UserSummary{ID: u.ID, Name: u.Name, Login: u.Login, Role: u.Role, CreatedAt: u.CreatedAt})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
