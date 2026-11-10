package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMissingOrMalformedAuthorization(t *testing.T) {
	auth := middleware.NewAuth(servicemocks.NewMockAuthService(t))
	handler := auth.RequireAuth(func(w http.ResponseWriter, _ *http.Request) { t.Fatal("request must not reach handler") })
	for _, header := range []string{"", "Basic abc", "Bearer", "Bearer a b"} {
		request := httptest.NewRequest("GET", "/?access_token=abc", nil)
		if header != "" {
			request.Header.Set("Authorization", header)
		}
		response := httptest.NewRecorder()
		handler(response, request)
		require.Equal(t, 401, response.Code)
		require.Equal(t, "Bearer", response.Header().Get("WWW-Authenticate"))
	}
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Add("Authorization", "Bearer a")
	request.Header.Add("Authorization", "Bearer b")
	response := httptest.NewRecorder()
	handler(response, request)
	require.Equal(t, 401, response.Code)
}

func TestSessionErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{service.ErrUnauthorized, 401}, {errors.New("database unavailable"), 500},
	} {
		authService := servicemocks.NewMockAuthService(t)
		authService.EXPECT().Authenticate(mock.Anything, "token").Return(model.Session{}, tc.err)
		handler := middleware.NewAuth(authService).RequireAuth(func(w http.ResponseWriter, _ *http.Request) { t.Fatal("request must not reach handler") })
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Set("Authorization", "Bearer token")
		response := httptest.NewRecorder()
		handler(response, request)
		require.Equal(t, tc.status, response.Code)
	}
}

func TestOwnerAndAdminAreDifferentPermissions(t *testing.T) {
	for _, tc := range []struct {
		role, path string
		status     int
	}{
		{model.RoleUser, "/users/1", 204},
		{model.RoleAdmin, "/users/1", 204},
		{model.RoleUser, "/users/2", 403},
		{model.RoleAdmin, "/users/2", 403},
		{model.RoleUser, "/users/abc", 400},
		{model.RoleUser, "/admin", 403},
		{model.RoleAdmin, "/admin", 204},
	} {
		t.Run(tc.role+tc.path, func(t *testing.T) {
			authService := servicemocks.NewMockAuthService(t)
			authService.EXPECT().Authenticate(mock.Anything, "token").Return(model.Session{UserID: 1, Role: tc.role}, nil)
			auth := middleware.NewAuth(authService)
			next := func(w http.ResponseWriter, r *http.Request) {
				session, ok := middleware.SessionFromContext(r.Context())
				require.True(t, ok)
				require.Equal(t, int64(1), session.UserID)
				w.WriteHeader(204)
			}
			mux := http.NewServeMux()
			mux.HandleFunc("GET /users/{user_id}", auth.RequireOwner(next))
			mux.HandleFunc("GET /admin", auth.RequireAdmin(next))
			request := httptest.NewRequest("GET", tc.path, nil)
			request.Header.Set("Authorization", "bearer token")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			require.Equal(t, tc.status, response.Code)
		})
	}
}
