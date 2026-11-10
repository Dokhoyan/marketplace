package user_test

import (
	"encoding/json"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers/user"
	"github.com/Dokhoyan/marketplace/internal/model"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	users := servicemocks.NewMockUserService(t)
	expected := model.User{ID: 1, Name: "Анна", Login: "anna", Role: "user", CreatedAt: time.Now().UTC()}
	users.EXPECT().CreateUser(mock.Anything, "Анна", "anna", "valid_password_123").Return(expected, nil)
	mux := http.NewServeMux()
	user.NewImplementation(users).RegisterRoutes(mux, middleware.NewAuth(nil))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"name":"Анна","login":"anna","password":"valid_password_123"}`)))
	require.Equal(t, http.StatusCreated, response.Code)
	var result model.User
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Equal(t, expected, result)
	require.NotContains(t, response.Body.String(), "password")
	require.NotContains(t, response.Body.String(), "valid_password_123")
}
