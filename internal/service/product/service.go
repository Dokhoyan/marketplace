package product

import (
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type serv struct {
	productRepo repository.ProductRepository
}

func NewService(productRepo repository.ProductRepository) service.ProductService {
	return &serv{productRepo: productRepo}
}
