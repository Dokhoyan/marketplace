package product

import (
	"encoding/json"
	"net/http"

	"github.com/Dokhoyan/marketplace/internal/logger"
	"github.com/Dokhoyan/marketplace/internal/model"
	"go.uber.org/zap"
)

func (i *Implementation) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := i.productService.GetProducts(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		logger.Error("failed to get products", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorBody{Code: model.ErrorInternal, Message: "внутренняя ошибка"},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(products)
}
