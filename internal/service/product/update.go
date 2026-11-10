package product

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) UpdateProduct(ctx context.Context, id int64, name *string, price *int64) (model.Product, error) {
	if name == nil && price == nil {
		return model.Product{}, service.ErrEmptyProductUpdate
	}
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if length := utf8.RuneCountInString(trimmedName); length < 1 || length > 200 {
			return model.Product{}, service.ErrInvalidProductName
		}
		name = &trimmedName
	}
	if price != nil && *price <= 0 {
		return model.Product{}, service.ErrInvalidPrice
	}

	product, err := s.productRepo.UpdateProduct(ctx, id, name, price)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return model.Product{}, service.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}
