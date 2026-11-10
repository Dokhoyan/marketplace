package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) SetProductStock(ctx context.Context, id, stock int64) (model.Product, error) {
	if stock < 0 {
		return model.Product{}, service.ErrInvalidStock
	}

	product, err := s.productRepo.SetProductStock(ctx, id, stock)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return model.Product{}, service.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to set product stock: %w", err)
	}

	return product, nil
}
