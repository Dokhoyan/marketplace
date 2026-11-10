package product_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers/product"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProductErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   model.ErrorCode
	}{
		{"not found", service.ErrProductNotFound, 404, model.ErrorNotFound},
		{"database error", errors.New("private database details"), 500, model.ErrorInternal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			products := servicemocks.NewMockProductService(t)
			products.EXPECT().GetProduct(mock.Anything, int64(1)).Return(model.Product{}, tc.err)
			mux := http.NewServeMux()
			product.NewImplementation(products).RegisterRoutes(mux, productAuth(t))
			response := productRequest(mux, "GET", "/api/v1/products/1", "")
			require.Equal(t, tc.status, response.Code)
			var result model.ErrorResponse
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			require.Equal(t, tc.code, result.Error.Code)
			require.NotContains(t, response.Body.String(), "private database details")
		})
	}
}
