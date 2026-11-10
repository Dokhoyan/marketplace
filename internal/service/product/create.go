package product

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) CreateProduct(ctx context.Context, name string, price, stock int64) (model.Product, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 1 || length > 200 {
		return model.Product{}, service.ErrInvalidProductName
	}
	if price <= 0 {
		return model.Product{}, service.ErrInvalidPrice
	}
	if stock < 0 {
		return model.Product{}, service.ErrInvalidStock
	}

	product, err := s.productRepo.CreateProduct(ctx, name, price, stock)
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to create product: %w", err)
	}
	return product, nil
}
