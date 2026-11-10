package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) AddProductStock(ctx context.Context, id, quantity int64) (model.Product, error) {
	if quantity <= 0 {
		return model.Product{}, service.ErrInvalidQuantity
	}

	product, err := s.productRepo.AddProductStock(ctx, id, quantity)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return model.Product{}, service.ErrProductNotFound
		}
		if errors.Is(err, repository.ErrStockOverflow) {
			return model.Product{}, service.ErrStockOverflow
		}
		return model.Product{}, fmt.Errorf("failed to add product stock: %w", err)
	}

	return product, nil
}
