package product_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	productImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/product"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/model"
	postgresRepo "github.com/Dokhoyan/marketplace/internal/repository/postgres"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	productserv "github.com/Dokhoyan/marketplace/internal/service/product"
	txmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

func productRequest(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer test-admin")
	mux.ServeHTTP(response, request)
	return response
}

func TestProductInvalidRequests(t *testing.T) {
	mux := http.NewServeMux()
	// Invalid requests must be rejected before reaching storage.
	productHandler := productImpl.NewImplementation(productserv.NewService(nil))
	productHandler.RegisterRoutes(mux, productAuth(t))
	cases := []struct{ method, path, body string }{
		{"GET", "/api/v1/products/abc", ""},
		{"GET", "/api/v1/products/0", ""},
		{"GET", "/api/v1/products/-1", ""},
		{"GET", "/api/v1/products/9223372036854775808", ""},
		{"DELETE", "/api/v1/products/no", ""},
		{"POST", "/api/v1/products", `{}`},
		{"POST", "/api/v1/products", `null`},
		{"POST", "/api/v1/products", `[]`},
		{"POST", "/api/v1/products", `{"name":"x","price":1}`},
		{"POST", "/api/v1/products", `{"name":" ","price":1,"stock":0}`},
		{"POST", "/api/v1/products", fmt.Sprintf(`{"name":%q,"price":1,"stock":0}`, strings.Repeat("я", 201))},
		{"POST", "/api/v1/products", `{"name":"x","price":0,"stock":0}`},
		{"POST", "/api/v1/products", `{"name":"x","price":-1,"stock":0}`},
		{"POST", "/api/v1/products", `{"name":"x","price":1,"stock":-1}`},
		{"POST", "/api/v1/products", `{"name":"x","price":1.5,"stock":0}`},
		{"POST", "/api/v1/products", `{"name":"x","price":9223372036854775808,"stock":0}`},
		{"POST", "/api/v1/products", `{"name":"x","price":"1","stock":0}`},
		{"POST", "/api/v1/products", `{"name":"x","price":1,"stock":null}`},
		{"POST", "/api/v1/products", `{"name":"x","price":1,"stock":0,"extra":true}`},
		{"POST", "/api/v1/products", `{"name":"x","price":1,"stock":0} {}`},
		{"PATCH", "/api/v1/products/1", `{}`},
		{"PATCH", "/api/v1/products/1", `{"name":null}`},
		{"PATCH", "/api/v1/products/1", `{"name":""}`},
		{"PATCH", "/api/v1/products/1", `{"price":0}`},
		{"PATCH", "/api/v1/products/1", `{"stock":2}`},
		{"PUT", "/api/v1/products/1/stock", `{}`},
		{"PUT", "/api/v1/products/1/stock", `{"stock":null}`},
		{"PUT", "/api/v1/products/1/stock", `{"stock":-1}`},
		{"POST", "/api/v1/products/1/stock/add", `{}`},
		{"POST", "/api/v1/products/1/stock/add", `{"quantity":0}`},
		{"POST", "/api/v1/products/1/stock/add", `{"quantity":-1}`},
	}
	for _, tc := range cases {
		t.Run(tc.method+tc.path+tc.body, func(t *testing.T) {
			response := productRequest(mux, tc.method, tc.path, tc.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestProductsPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("products_test_%d", time.Now().UnixNano())
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migration, err := os.ReadFile("../../../../../migrations/20261110000000_create_products.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.Split(string(migration), "-- +goose Down")[0]
	if _, err := pool.Exec(ctx, up); err != nil {
		t.Fatal(err)
	}
	newMux := func() *http.ServeMux {
		mux := http.NewServeMux()
		productHandler := productImpl.NewImplementation(productserv.NewService(postgresRepo.New(pool, txmpgx.DefaultCtxGetter)))
		productHandler.RegisterRoutes(mux, productAuth(t))
		return mux
	}
	mux := newMux()
	request := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		response := productRequest(mux, method, path, body)
		if response.Code != status {
			t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, status, response.Body.String())
		}
		return response
	}
	decode := func(response *httptest.ResponseRecorder) model.Product {
		t.Helper()
		var p model.Product
		if err := json.Unmarshal(response.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	assertProduct := func(p model.Product, name string, price, stock int64) {
		t.Helper()
		if p.ID <= 0 || p.Name != name || p.Price != price || p.Stock != stock {
			t.Fatalf("unexpected product: %+v", p)
		}
	}
	if body := request("GET", "/api/v1/products", "", 200).Body.String(); strings.TrimSpace(body) != "[]" {
		t.Fatalf("empty catalog = %s", body)
	}
	created := decode(request("POST", "/api/v1/products", `{"name":" Клавиатура ","price":20000,"stock":5}`, 201))
	assertProduct(created, "Клавиатура", 20000, 5)
	path := fmt.Sprintf("/api/v1/products/%d", created.ID)
	// A new service/repository instance reads the same persisted product.
	mux = newMux()
	assertProduct(decode(request("GET", path, "", 200)), "Клавиатура", 20000, 5)
	var catalog []model.Product
	if err := json.Unmarshal(request("GET", "/api/v1/products", "", 200).Body.Bytes(), &catalog); err != nil || len(catalog) != 1 || catalog[0] != created {
		t.Fatalf("catalog = %+v, err = %v", catalog, err)
	}
	assertProduct(decode(request("PATCH", path, `{"name":"Мышь"}`, 200)), "Мышь", 20000, 5)
	assertProduct(decode(request("PATCH", path, `{"price":10000}`, 200)), "Мышь", 10000, 5)
	assertProduct(decode(request("PATCH", path, `{"name":"Монитор","price":30000}`, 200)), "Монитор", 30000, 5)
	assertProduct(decode(request("PUT", path+"/stock", `{"stock":0}`, 200)), "Монитор", 30000, 0)
	assertProduct(decode(request("POST", path+"/stock/add", `{"quantity":3}`, 200)), "Монитор", 30000, 3)

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := productRequest(mux, "POST", path+"/stock/add", `{"quantity":1}`)
			if response.Code != 200 {
				t.Errorf("concurrent add: %d %s", response.Code, response.Body.String())
			}
		}()
	}
	wg.Wait()
	assertProduct(decode(request("GET", path, "", 200)), "Монитор", 30000, 23)
	request("PUT", path+"/stock", `{"stock":9223372036854775807}`, 200)
	request("POST", path+"/stock/add", `{"quantity":1}`, 400)
	assertProduct(decode(request("GET", path, "", 200)), "Монитор", 30000, 9223372036854775807)

	// Model an existing order reference without adding order functionality.
	if _, err := pool.Exec(ctx, `CREATE TABLE order_reference (product_id BIGINT REFERENCES products(id))`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO order_reference (product_id) VALUES ($1)`, created.ID); err != nil {
		t.Fatal(err)
	}
	if response := request("DELETE", path, "", 204); response.Body.Len() != 0 {
		t.Fatalf("DELETE body = %s", response.Body.String())
	}
	for _, tc := range []struct{ method, suffix, body string }{
		{"GET", "", ""}, {"PATCH", "", `{"price":1}`},
		{"PUT", "/stock", `{"stock":1}`}, {"POST", "/stock/add", `{"quantity":1}`}, {"DELETE", "", ""},
	} {
		request(tc.method, path+tc.suffix, tc.body, 404)
		request(tc.method, "/api/v1/products/9223372036854775807"+tc.suffix, tc.body, 404)
	}
	if body := request("GET", "/api/v1/products", "", 200).Body.String(); strings.TrimSpace(body) != "[]" {
		t.Fatalf("deleted product in catalog: %s", body)
	}
	var preserved bool
	if err := pool.QueryRow(ctx, `SELECT p.deleted_at IS NOT NULL FROM order_reference o JOIN products p ON p.id = o.product_id`).Scan(&preserved); err != nil || !preserved {
		t.Fatalf("order reference not preserved: %v", err)
	}
	// UTF-8 name length is counted in characters, not bytes.
	name := strings.Repeat("я", 200)
	assertProduct(decode(request("POST", "/api/v1/products", fmt.Sprintf(`{"name":%q,"price":1,"stock":0}`, name), 201)), name, 1, 0)
	if _, err := pool.Exec(ctx, `DROP TABLE order_reference`); err != nil {
		t.Fatal(err)
	}
	down := strings.Split(string(migration), "-- +goose Down")[1]
	if _, err := pool.Exec(ctx, down); err != nil {
		t.Fatalf("down migration: %v", err)
	}
}

func productAuth(t *testing.T) *middleware.Auth {
	auth := servicemocks.NewMockAuthService(t)
	auth.EXPECT().Authenticate(mock.Anything, "test-admin").Return(model.Session{UserID: 1, Role: model.RoleAdmin}, nil).Maybe()
	return middleware.NewAuth(auth)
}
