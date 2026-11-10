package auth

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

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (i *Implementation) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	var request LoginRequest
	if err := handlers.DecodeJSON(r, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный JSON"}})
		return
	}
	token, err := i.authService.Login(r.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorUnauthorized, Message: "Неверный логин или пароль"}})
			return
		}
		logger.Error("failed to login", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"}})
		return
	}
	w.WriteHeader(http.StatusOK)
	// #nosec G117 -- access_token возвращается клиенту по контракту входа; хеши и пароли сюда не попадают.
	_ = json.NewEncoder(w).Encode(token)
}
