package product

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

type CreateProductRequest struct {
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Stock *int64 `json:"stock"`
}

func (i *Implementation) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var request CreateProductRequest
	if err := handlers.DecodeJSON(r, &request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный JSON"},
		})
		return
	}

	if request.Stock == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "передайте целочисленное поле stock"},
		})
		return
	}

	product, err := i.productService.CreateProduct(r.Context(), request.Name, request.Price, *request.Stock)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProductName) ||
			errors.Is(err, service.ErrInvalidPrice) ||
			errors.Is(err, service.ErrInvalidStock) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: err.Error()},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		logger.Error("failed to create product", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(product)
}
