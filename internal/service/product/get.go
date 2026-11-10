package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) GetProduct(ctx context.Context, id int64) (model.Product, error) {
	product, err := s.productRepo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return model.Product{}, service.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}
