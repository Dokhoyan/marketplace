package user_test

import (
	"encoding/json"
	"errors"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/user"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	userserv "github.com/Dokhoyan/marketplace/internal/service/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegistrationInvalidRequests(t *testing.T) {
	mux := http.NewServeMux()
	userImpl.NewImplementation(userserv.NewService(nil, nil, nil, nil)).RegisterRoutes(mux, middleware.NewAuth(nil))
	for _, body := range []string{
		`{}`, `null`, `[]`,
		`{"name":"Анна"}`,
		`{"name":"Анна","login":null,"password":"valid_password_123"}`,
		`{"name":"Анна","login":"anna","password":null}`,
		`{"name":"Анна","login":123,"password":"valid_password_123"}`,
		`{"name":"Анна","login":"anna","password":"short"}`,
		`{"name":"Анна","login":"anna","password":"valid_password_123","role":"admin"}`,
		`{"name":"Анна","login":"anna","password":"valid_password_123","balance":100}`,
		`{"name":"Анна","login":"anna","password":"valid_password_123"} {}`,
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(body)))
		require.Equal(t, http.StatusBadRequest, response.Code)
	}
}

func TestRegistrationErrors(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   model.ErrorCode
	}{
		{service.ErrLoginTaken, 409, model.ErrorLoginTaken},
		{errors.New("private database details"), 500, model.ErrorInternal},
	} {
		users := servicemocks.NewMockUserService(t)
		users.EXPECT().CreateUser(mock.Anything, "Анна", "anna", "valid_password_123").Return(model.User{}, tc.err)
		mux := http.NewServeMux()
		userImpl.NewImplementation(users).RegisterRoutes(mux, middleware.NewAuth(nil))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"name":"Анна","login":"anna","password":"valid_password_123"}`)))
		require.Equal(t, tc.status, response.Code)
		var result model.ErrorResponse
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
		require.Equal(t, tc.code, result.Error.Code)
		require.NotContains(t, response.Body.String(), "private database details")
	}
}
