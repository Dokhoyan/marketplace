package product_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	repomocks "github.com/Dokhoyan/marketplace/internal/repository/mocks"
	"github.com/Dokhoyan/marketplace/internal/service"
	"github.com/Dokhoyan/marketplace/internal/service/product"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()
	repo := repomocks.NewMockProductRepository(t)
	expected := model.Product{ID: 1, Name: "Клавиатура", Price: 20000, Stock: 5}
	repo.EXPECT().CreateProduct(ctx, "Клавиатура", int64(20000), int64(5)).Return(expected, nil)

	result, err := product.NewService(repo).CreateProduct(ctx, " Клавиатура ", 20000, 5)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestUpdateProductPreservesMissingFields(t *testing.T) {
	ctx := context.Background()
	repo := repomocks.NewMockProductRepository(t)
	name := " Мышь "
	trimmed := "Мышь"
	expected := model.Product{ID: 1, Name: trimmed, Price: 20000, Stock: 5}
	repo.EXPECT().UpdateProduct(ctx, int64(1), &trimmed, (*int64)(nil)).Return(expected, nil)

	result, err := product.NewService(repo).UpdateProduct(ctx, 1, &name, nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.Equal(t, " Мышь ", name)
}

func TestProductRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		setup func(*repomocks.MockProductRepository)
		call  func(service.ProductService) error
		want  error
	}{
		{
			name: "get missing product",
			setup: func(repo *repomocks.MockProductRepository) {
				repo.EXPECT().GetProduct(ctx, int64(1)).Return(model.Product{}, repository.ErrProductNotFound)
			},
			call: func(s service.ProductService) error { _, err := s.GetProduct(ctx, 1); return err },
			want: service.ErrProductNotFound,
		},
		{
			name: "update missing product",
			setup: func(repo *repomocks.MockProductRepository) {
				price := int64(100)
				repo.EXPECT().UpdateProduct(ctx, int64(1), (*string)(nil), &price).Return(model.Product{}, repository.ErrProductNotFound)
			},
			call: func(s service.ProductService) error {
				price := int64(100)
				_, err := s.UpdateProduct(ctx, 1, nil, &price)
				return err
			},
			want: service.ErrProductNotFound,
		},
		{
			name: "set stock of missing product",
			setup: func(repo *repomocks.MockProductRepository) {
				repo.EXPECT().SetProductStock(ctx, int64(1), int64(0)).Return(model.Product{}, repository.ErrProductNotFound)
			},
			call: func(s service.ProductService) error { _, err := s.SetProductStock(ctx, 1, 0); return err },
			want: service.ErrProductNotFound,
		},
		{
			name: "add stock to missing product",
			setup: func(repo *repomocks.MockProductRepository) {
				repo.EXPECT().AddProductStock(ctx, int64(1), int64(1)).Return(model.Product{}, repository.ErrProductNotFound)
			},
			call: func(s service.ProductService) error { _, err := s.AddProductStock(ctx, 1, 1); return err },
			want: service.ErrProductNotFound,
		},
		{
			name: "stock overflow",
			setup: func(repo *repomocks.MockProductRepository) {
				repo.EXPECT().AddProductStock(ctx, int64(1), int64(1)).Return(model.Product{}, repository.ErrStockOverflow)
			},
			call: func(s service.ProductService) error { _, err := s.AddProductStock(ctx, 1, 1); return err },
			want: service.ErrStockOverflow,
		},
		{
			name: "delete missing product",
			setup: func(repo *repomocks.MockProductRepository) {
				repo.EXPECT().DeleteProduct(ctx, int64(1)).Return(repository.ErrProductNotFound)
			},
			call: func(s service.ProductService) error { return s.DeleteProduct(ctx, 1) },
			want: service.ErrProductNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := repomocks.NewMockProductRepository(t)
			tc.setup(repo)
			require.ErrorIs(t, tc.call(product.NewService(repo)), tc.want)
		})
	}
}

func TestProductDatabaseError(t *testing.T) {
	ctx := context.Background()
	repo := repomocks.NewMockProductRepository(t)
	dbErr := errors.New("database unavailable")
	repo.EXPECT().GetProducts(ctx).Return(nil, dbErr)
	_, err := product.NewService(repo).GetProducts(ctx)
	require.ErrorIs(t, err, dbErr)
	require.ErrorContains(t, err, "failed to get products")
}
