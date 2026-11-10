package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

type sessionContextKey struct{}

type Auth struct {
	authService service.AuthService
}

func NewAuth(authService service.AuthService) *Auth {
	return &Auth{authService: authService}
}

func SessionFromContext(ctx context.Context) (model.Session, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(model.Session)
	return session, ok
}

func writeAccessError(w http.ResponseWriter, status int, code model.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: model.ErrorBody{Code: code, Message: message}})
}

func (a *Auth) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fields := strings.Fields(r.Header.Get("Authorization"))
		if len(r.Header.Values("Authorization")) != 1 || len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			writeAccessError(w, http.StatusUnauthorized, model.ErrorUnauthorized, service.ErrUnauthorized.Error())
			return
		}
		session, err := a.authService.Authenticate(r.Context(), fields[1])
		if err != nil {
			if errors.Is(err, service.ErrUnauthorized) {
				writeAccessError(w, http.StatusUnauthorized, model.ErrorUnauthorized, err.Error())
			} else {
				logger.Error("failed to authenticate session", zap.Error(err))
				writeAccessError(w, http.StatusInternalServerError, model.ErrorInternal, "внутренняя ошибка")
			}
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), sessionContextKey{}, session)))
	}
}

func (a *Auth) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		session, _ := SessionFromContext(r.Context())
		if session.Role != model.RoleAdmin {
			writeAccessError(w, http.StatusForbidden, model.ErrorForbidden, "Недостаточно прав")
			return
		}
		next(w, r)
	})
}

// У администратора нет исключения из проверки владельца.
func (a *Auth) RequireOwner(next http.HandlerFunc) http.HandlerFunc {
	return a.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
		if err != nil || id <= 0 {
			writeAccessError(w, http.StatusBadRequest, model.ErrorBadRequest, "ID пользователя должен быть положительным целым числом")
			return
		}
		session, _ := SessionFromContext(r.Context())
		if id != session.UserID {
			writeAccessError(w, http.StatusForbidden, model.ErrorForbidden, "Доступ разрешён только владельцу")
			return
		}
		next(w, r)
	})
}
