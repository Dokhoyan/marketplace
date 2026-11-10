package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"go.uber.org/zap"
)

func (i *Implementation) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := middleware.SessionFromContext(r.Context())
	if err := i.authService.Logout(r.Context(), session.TokenHash); err != nil {
		logger.Error("failed to logout", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"}})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
