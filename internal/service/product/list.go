package product

import (
	"context"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
)

func (s *serv) GetProducts(ctx context.Context) ([]model.Product, error) {
	products, err := s.productRepo.GetProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	return products, nil
}
