package product

import (
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type Implementation struct {
	productService service.ProductService
}

func NewImplementation(s service.ProductService) *Implementation {
	return &Implementation{productService: s}
}

func (i *Implementation) RegisterRoutes(mux *http.ServeMux, auth *middleware.Auth) {
	mux.HandleFunc("GET /api/v1/products", i.GetProducts)
	mux.HandleFunc("GET /api/v1/products/{product_id}", i.GetProduct)
	mux.HandleFunc("POST /api/v1/products", auth.RequireAdmin(i.CreateProduct))
	mux.HandleFunc("PATCH /api/v1/products/{product_id}", auth.RequireAdmin(i.UpdateProduct))
	mux.HandleFunc("PUT /api/v1/products/{product_id}/stock", auth.RequireAdmin(i.SetProductStock))
	mux.HandleFunc("POST /api/v1/products/{product_id}/stock/add", auth.RequireAdmin(i.AddProductStock))
	mux.HandleFunc("DELETE /api/v1/products/{product_id}", auth.RequireAdmin(i.DeleteProduct))
}
