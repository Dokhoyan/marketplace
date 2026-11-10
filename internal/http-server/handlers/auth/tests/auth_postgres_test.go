package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/auth"
	productImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/product"
	userImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/user"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/model"
	postgresRepo "github.com/Dokhoyan/marketplace/internal/repository/postgres"
	authserv "github.com/Dokhoyan/marketplace/internal/service/auth"
	productserv "github.com/Dokhoyan/marketplace/internal/service/product"
	userserv "github.com/Dokhoyan/marketplace/internal/service/user"
	txmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestAuthPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	adminPool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer adminPool.Close()
	schema := pgx.Identifier{fmt.Sprintf("auth_test_%d", time.Now().UnixNano())}.Sanitize()
	_, err = adminPool.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() {
		_, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		require.NoError(t, err)
	}()
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	migrations, err := filepath.Glob("../../../../../migrations/*.sql")
	require.NoError(t, err)
	for _, path := range migrations {
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, strings.Split(string(data), "-- +goose Down")[0])
		require.NoError(t, err)
	}
	repo := postgresRepo.New(pool, txmpgx.DefaultCtxGetter)
	users := userserv.NewService(repo, repo, repo, manager.Must(txmpgx.NewDefaultFactory(pool)))
	newMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		// New instances use the same persisted sessions after a server restart.
		service := authserv.NewService(repo, repo)
		auth := middleware.NewAuth(service)
		authImpl.NewImplementation(service).RegisterRoutes(mux, auth)
		userImpl.NewImplementation(users).RegisterRoutes(mux, auth)
		productImpl.NewImplementation(productserv.NewService(repo)).RegisterRoutes(mux, auth)
		return mux
	}
	mux := newMux()
	request := func(method, path, token, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		require.Equal(t, status, response.Code, "%s %s: %s", method, path, response.Body.String())
		return response
	}
	password := "valid_password_123"
	register := func(login string) model.User {
		t.Helper()
		body, err := json.Marshal(map[string]string{"name": "Имя", "login": login, "password": password})
		require.NoError(t, err)
		var created model.User
		require.NoError(t, json.Unmarshal(request("POST", "/api/v1/users", "", string(body), 201).Body.Bytes(), &created))
		return created
	}
	login := func(name string) model.AccessToken {
		t.Helper()
		body, err := json.Marshal(map[string]string{"login": name, "password": password})
		require.NoError(t, err)
		response := request("POST", "/api/v1/auth/login", "", string(body), 200)
		require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
		var token model.AccessToken
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &token))
		require.Equal(t, "Bearer", token.TokenType)
		require.Len(t, token.AccessToken, 43)
		return token
	}
	alice := register("alice")
	bob := register("bob")
	administrator := register("admin")
	// Администратор в фикстуре — обычный пользователь с назначенной через БД ролью.
	_, err = pool.Exec(ctx, `UPDATE users SET role='admin' WHERE id=$1`, administrator.ID)
	require.NoError(t, err)
	administrator, err = repo.GetUser(ctx, administrator.ID)
	require.NoError(t, err)
	require.Equal(t, model.RoleAdmin, administrator.Role)
	require.Zero(t, administrator.Balance)
	var adminCarts int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM carts WHERE user_id=$1`, administrator.ID).Scan(&adminCarts))
	require.Equal(t, 1, adminCarts)
	alicePath := fmt.Sprintf("/api/v1/users/%d", alice.ID)
	bobPath := fmt.Sprintf("/api/v1/users/%d", bob.ID)
	adminPath := fmt.Sprintf("/api/v1/users/%d", administrator.ID)
	first := login("ALICE")
	second := login("alice")
	require.NotEqual(t, first.AccessToken, second.AccessToken)
	adminToken := login("admin").AccessToken
	bobToken := login("bob").AccessToken
	hash := sha256.Sum256([]byte(first.AccessToken))
	firstHash := hex.EncodeToString(hash[:])
	var duration float64
	var storedHash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT token_hash, extract(epoch FROM expires_at-created_at) FROM sessions WHERE token_hash=$1`, firstHash).Scan(&storedHash, &duration))
	require.Equal(t, firstHash, storedHash)
	require.NotEqual(t, first.AccessToken, storedHash)
	require.Equal(t, float64(24*60*60), duration)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), first.ExpiresAt, 10*time.Second)
	mux = newMux()
	profile := request("GET", alicePath, first.AccessToken, "", 200)
	require.NotContains(t, profile.Body.String(), "password")
	require.NotContains(t, profile.Body.String(), "hash")
	request("GET", alicePath, "", "", 401)
	request("GET", alicePath, strings.Repeat("x", 43), "", 401)
	for _, token := range []string{bobToken, adminToken} {
		request("GET", alicePath, token, "", 403)
		request("PATCH", alicePath, token, `{"name":"Чужое"}`, 403)
		request("DELETE", alicePath, token, "", 403)
	}
	request("PATCH", alicePath, first.AccessToken, `{"name":"Новое имя"}`, 200)
	request("PATCH", alicePath, first.AccessToken, `{"name":" "}`, 400)
	request("PATCH", alicePath, first.AccessToken, `{"name":"Имя","role":"admin"}`, 400)
	request("PATCH", alicePath, first.AccessToken, `{"balance":100}`, 400)
	request("GET", "/api/v1/users", first.AccessToken, "", 403)
	request("GET", "/api/v1/users", "", "", 401)
	listing := request("GET", "/api/v1/users", adminToken, "", 200)
	require.NotContains(t, listing.Body.String(), "balance")
	require.NotContains(t, listing.Body.String(), "password")
	request("GET", adminPath, adminToken, "", 200)
	request("GET", bobPath, bobToken, "", 200)
	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/api/v1/products", `{"name":"Товар","price":100,"stock":1}`},
		{"PATCH", "/api/v1/products/1", `{"price":200}`},
		{"PUT", "/api/v1/products/1/stock", `{"stock":5}`},
		{"POST", "/api/v1/products/1/stock/add", `{"quantity":2}`},
		{"DELETE", "/api/v1/products/1", ""},
	} {
		request(tc.method, tc.path, "", tc.body, 401)
		request(tc.method, tc.path, first.AccessToken, tc.body, 403)
	}
	product := request("POST", "/api/v1/products", adminToken, `{"name":"Товар","price":100,"stock":1}`, 201)
	var createdProduct model.Product
	require.NoError(t, json.Unmarshal(product.Body.Bytes(), &createdProduct))
	productPath := fmt.Sprintf("/api/v1/products/%d", createdProduct.ID)
	request("GET", "/api/v1/products", "", "", 200)
	request("GET", productPath, "", "", 200)
	request("PATCH", productPath, adminToken, `{"price":200}`, 200)
	request("PUT", productPath+"/stock", adminToken, `{"stock":5}`, 200)
	request("POST", productPath+"/stock/add", adminToken, `{"quantity":2}`, 200)
	request("DELETE", productPath, adminToken, "", 204)
	response := request("POST", "/api/v1/auth/logout", first.AccessToken, "", 204)
	require.Empty(t, response.Body.String())
	request("GET", alicePath, first.AccessToken, "", 401)
	request("POST", "/api/v1/auth/logout", first.AccessToken, "", 401)
	request("GET", alicePath, second.AccessToken, "", 200)
	// Expiration is checked by PostgreSQL rather than the application clock.
	_, err = pool.Exec(ctx, `UPDATE sessions SET expires_at=now()-interval '1 second' WHERE user_id=$1`, alice.ID)
	require.NoError(t, err)
	request("GET", alicePath, second.AccessToken, "", 401)
	third := login("alice")
	fourth := login("alice")
	wrongLogin := request("POST", "/api/v1/auth/login", "", `{"login":"missing","password":"valid_password_123"}`, 401).Body.String()
	wrongPassword := request("POST", "/api/v1/auth/login", "", `{"login":"alice","password":"wrong_password_123"}`, 401).Body.String()
	require.Equal(t, wrongLogin, wrongPassword)
	require.Contains(t, wrongLogin, "Неверный логин или пароль")
	// Role is read on each request, so a removed admin privilege takes effect immediately.
	_, err = pool.Exec(ctx, `UPDATE users SET role='user' WHERE id=$1`, administrator.ID)
	require.NoError(t, err)
	request("GET", "/api/v1/users", adminToken, "", 403)
	_, err = pool.Exec(ctx, `UPDATE users SET role='admin' WHERE id=$1`, administrator.ID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
 CREATE FUNCTION reject_session_delete() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN RAISE EXCEPTION 'session deletion failed'; END;
 $$;
 CREATE TRIGGER reject_session_delete BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION reject_session_delete();
 `)
	require.NoError(t, err)
	request("DELETE", alicePath, third.AccessToken, "", 500)
	request("GET", alicePath, third.AccessToken, "", 200)
	_, err = pool.Exec(ctx, `DROP TRIGGER reject_session_delete ON sessions; DROP FUNCTION reject_session_delete()`)
	require.NoError(t, err)
	require.Empty(t, request("DELETE", alicePath, third.AccessToken, "", 204).Body.String())
	request("GET", alicePath, third.AccessToken, "", 401)
	request("GET", alicePath, fourth.AccessToken, "", 401)
	deletedLogin := request("POST", "/api/v1/auth/login", "", `{"login":"alice","password":"valid_password_123"}`, 401).Body.String()
	require.Equal(t, wrongLogin, deletedLogin)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM sessions WHERE user_id=$1`, alice.ID).Scan(&count))
	require.Zero(t, count)
	var deleted bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT deleted_at IS NOT NULL FROM users WHERE id=$1`, alice.ID).Scan(&deleted))
	require.True(t, deleted)
	request("POST", "/api/v1/users", "", `{"name":"Снова","login":"alice","password":"valid_password_123"}`, 409)
	listing = request("GET", "/api/v1/users", adminToken, "", 200)
	require.NotContains(t, listing.Body.String(), `"login":"alice"`)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM carts WHERE user_id=$1`, alice.ID).Scan(&count))
	require.Equal(t, 1, count)
	// Полный откат и повторное применение проверяют порядок зависимостей таблиц.
	for i := len(migrations) - 1; i >= 0; i-- {
		data, err := os.ReadFile(migrations[i])
		require.NoError(t, err)
		_, err = pool.Exec(ctx, strings.Split(string(data), "-- +goose Down")[1])
		require.NoError(t, err)
	}
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema()`).Scan(&count))
	require.Zero(t, count)
	for _, path := range migrations {
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, strings.Split(string(data), "-- +goose Down")[0])
		require.NoError(t, err)
	}
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema()`).Scan(&count))
	require.Equal(t, 5, count)
}
