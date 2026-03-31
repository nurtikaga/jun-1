package app

import (
	"context"
	"fmt"

	"github.com/nurtikaga/jun-1/internal/domain/product"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id product.ProductID) (*product.Product, error)
	Save(ctx context.Context, p *product.Product) error
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, name string, priceInCents int64, stock int) (*product.Product, error) {
	p, err := product.New(product.NewProductID(), name, priceInCents, stock, product.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("app: create product: %w", err)
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("app: save product: %w", err)
	}
	return p, nil
}

func (s *ProductService) UpdatePrice(ctx context.Context, id product.ProductID, newPrice int64) (*product.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("app: find product: %w", err)
	}
	if err := p.UpdatePrice(newPrice); err != nil {
		return nil, fmt.Errorf("app: update price: %w", err)
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("app: save product: %w", err)
	}
	return p, nil
}

func (s *ProductService) DecreaseStock(ctx context.Context, id product.ProductID, quantity int) (*product.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("app: find product: %w", err)
	}
	if err := p.DecreaseStock(quantity); err != nil {
		return nil, fmt.Errorf("app: decrease stock: %w", err)
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("app: save product: %w", err)
	}
	return p, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id product.ProductID) (*product.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("app: find product: %w", err)
	}
	return p, nil
}
