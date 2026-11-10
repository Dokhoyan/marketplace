package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) DeleteProduct(ctx context.Context, id int64) error {
	err := s.productRepo.DeleteProduct(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return service.ErrProductNotFound
		}
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}
