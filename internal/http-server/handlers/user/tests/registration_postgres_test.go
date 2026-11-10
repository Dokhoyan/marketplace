package user_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	userImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/user"
	"github.com/Dokhoyan/marketplace/internal/model"
	postgresRepo "github.com/Dokhoyan/marketplace/internal/repository/postgres"
	userserv "github.com/Dokhoyan/marketplace/internal/service/user"
	txmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestRegistrationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer admin.Close()
	schema := pgx.Identifier{fmt.Sprintf("registration_test_%d", time.Now().UnixNano())}.Sanitize()
	_, err = admin.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() {
		_, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		require.NoError(t, err)
	}()
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	var registrationMigration string
	for _, name := range []string{"20250930120000_create_users.sql", "20261110000000_create_products.sql", "20261110010000_create_carts.sql"} {
		migration, err := os.ReadFile("../../../../../migrations/" + name)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, strings.Split(string(migration), "-- +goose Down")[0])
		require.NoError(t, err)
		if name == "20261110010000_create_carts.sql" {
			registrationMigration = string(migration)
		}
	}
	mux := http.NewServeMux()
	repo := postgresRepo.New(pool, txmpgx.DefaultCtxGetter)
	userImpl.NewImplementation(userserv.NewService(repo, repo, repo, manager.Must(txmpgx.NewDefaultFactory(pool)))).RegisterRoutes(mux, middleware.NewAuth(nil))
	request := func(name, login, password string) *httptest.ResponseRecorder {
		payload, err := json.Marshal(map[string]string{"name": name, "login": login, "password": password})
		require.NoError(t, err)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(string(payload))))
		return response
	}
	password := "valid_password_123"
	response := request(" Анна ", "Anna_1", password)
	require.Equal(t, 201, response.Code, response.Body.String())
	var created model.User
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &created))
	require.Positive(t, created.ID)
	require.Equal(t, "Анна", created.Name)
	require.Equal(t, "anna_1", created.Login)
	require.Equal(t, "user", created.Role)
	require.Zero(t, created.Balance)
	require.False(t, created.CreatedAt.IsZero())
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &fields))
	require.Len(t, fields, 6)
	require.NotContains(t, fields, "password")
	require.NotContains(t, fields, "password_hash")
	require.NotContains(t, response.Body.String(), password)

	var stored model.User
	var encoded string
	require.NoError(t, pool.QueryRow(ctx, `SELECT id, name, login, role, balance, created_at, password_hash FROM users WHERE id = $1`, created.ID).Scan(
		&stored.ID, &stored.Name, &stored.Login, &stored.Role, &stored.Balance, &stored.CreatedAt, &encoded,
	))
	require.Equal(t, created.ID, stored.ID)
	require.Equal(t, created.Login, stored.Login)
	require.Equal(t, "user", stored.Role)
	require.Zero(t, stored.Balance)
	parts := strings.Split(encoded, "$")
	require.Len(t, parts, 6)
	require.Equal(t, "argon2id", parts[1])
	var memory, iterations uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads)
	require.NoError(t, err)
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	require.NoError(t, err)
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	require.NoError(t, err)
	require.Equal(t, hash, argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(hash))))
	var carts, items int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM carts WHERE user_id = $1`, created.ID).Scan(&carts))
	require.Equal(t, 1, carts)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM cart_items ci JOIN carts c ON c.id = ci.cart_id WHERE c.user_id = $1`, created.ID).Scan(&items))
	require.Zero(t, items)
	require.Equal(t, 409, request("Другой", "ANNA_1", password).Code)
	// The unique index enforces case-insensitive uniqueness even outside the service.
	_, err = pool.Exec(ctx, `INSERT INTO users (name, login, password_hash) VALUES ('Другой', 'ANNA_1', $1)`, encoded)
	require.Error(t, err)

	// Concurrent registration of the same login creates exactly one account.
	responses := make(chan int, 2)
	var wg sync.WaitGroup
	for _, login := range []string{"parallel", "PARALLEL"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			responses <- request("Параллельный", login, password).Code
		}()
	}
	wg.Wait()
	close(responses)
	var statuses []int
	for status := range responses {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	require.Equal(t, []int{201, 409}, statuses)

	// If cart creation fails, the user insert must be rolled back too.
	_, err = pool.Exec(ctx, `
  CREATE FUNCTION reject_cart() RETURNS trigger LANGUAGE plpgsql AS $$
  BEGIN RAISE EXCEPTION 'cart creation failed'; END;
  $$;
  CREATE TRIGGER reject_cart BEFORE INSERT ON carts FOR EACH ROW EXECUTE FUNCTION reject_cart();
 `)
	require.NoError(t, err)
	require.Equal(t, 500, request("Ошибка корзины", "rollback_user", password).Code)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE login = 'rollback_user'`).Scan(&count))
	require.Zero(t, count)
	_, err = pool.Exec(ctx, `DROP TRIGGER reject_cart ON carts; DROP FUNCTION reject_cart()`)
	require.NoError(t, err)

	// Character boundaries, including passwords longer than 72 bytes.
	require.Equal(t, 201, request(strings.Repeat("я", 200), "a_1", strings.Repeat("я", 15)).Code)
	require.Equal(t, 201, request("Максимум", strings.Repeat("a", 32), strings.Repeat("я", 128)).Code)
	_, err = pool.Exec(ctx, strings.Split(registrationMigration, "-- +goose Down")[1])
	require.NoError(t, err)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count))
	require.Equal(t, 4, count)
}
