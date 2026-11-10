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

type SetProductStockRequest struct {
	Stock *int64 `json:"stock"`
}

func (i *Implementation) SetProductStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorBadRequest, Message: "ID товара должен быть положительным целым числом"},
		})
		return
	}

	var request SetProductStockRequest
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

	product, err := i.productService.SetProductStock(r.Context(), id, *request.Stock)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStock) {
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
		logger.Error("failed to set product stock", zap.Error(err))
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
