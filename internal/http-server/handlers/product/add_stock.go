package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Dokhoyan/marketplace/internal/http-server/handlers"
	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
	"go.uber.org/zap"
)

type AddProductStockRequest struct {
	Quantity int64 `json:"quantity"`
}

func (i *Implementation) AddProductStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "ID товара должен быть положительным целым числом"},
		})
		return
	}

	var request AddProductStockRequest
	if err := handlers.DecodeJSON(r, &request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "некорректный JSON"},
		})
		return
	}

	product, err := i.productService.AddProductStock(r.Context(), id, request.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidQuantity) || errors.Is(err, service.ErrStockOverflow) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: err.Error()},
			})
			return
		}

		if errors.Is(err, service.ErrProductNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(model.ErrorResponse{
				Error: model.ErrorBody{Code: model.ErrorNotFound, Message: err.Error()},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		logger.Error("failed to add product stock", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(product)
}
