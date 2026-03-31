package app_test

import (
	"context"
	"testing"

	"github.com/nurtikaga/jun-1/internal/app"
	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
	"github.com/nurtikaga/jun-1/internal/domain/product"
)


type memRepo struct {
	data map[product.ProductID]*product.Product
}

func newMemRepo() *memRepo {
	return &memRepo{data: make(map[product.ProductID]*product.Product)}
}

func (r *memRepo) FindByID(_ context.Context, id product.ProductID) (*product.Product, error) {
	p, ok := r.data[id]
	if !ok {
		return nil, domainerrors.ErrProductNotFound
	}
	return p, nil
}

func (r *memRepo) Save(_ context.Context, p *product.Product) error {
	r.data[p.ID()] = p
	return nil
}


func setup(t *testing.T) (*app.ProductService, *memRepo) {
	t.Helper()
	repo := newMemRepo()
	svc := app.NewProductService(repo)
	return svc, repo
}


func TestCreateProduct_Success(t *testing.T) {
	svc, _ := setup(t)
	ctx := context.Background()

	p, err := svc.CreateProduct(ctx, "Gadget", 999, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "Gadget" {
		t.Errorf("want name=Gadget, got %s", p.Name())
	}
	if p.Price() != 999 {
		t.Errorf("want price=999, got %d", p.Price())
	}
}

func TestCreateProduct_NegativePrice(t *testing.T) {
	svc, _ := setup(t)
	_, err := svc.CreateProduct(context.Background(), "Bad", -1, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !domainerrors.Is(err, domainerrors.ErrInvalidPrice) {
		t.Errorf("want ErrInvalidPrice, got %v", err)
	}
}

func TestUpdatePrice_NotFound(t *testing.T) {
	svc, _ := setup(t)
	_, err := svc.UpdatePrice(context.Background(), product.ProductID("ghost"), 100)
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestDecreaseStock_Success(t *testing.T) {
	svc, _ := setup(t)
	ctx := context.Background()

	p, _ := svc.CreateProduct(ctx, "Widget", 500, 10)
	updated, err := svc.DecreaseStock(ctx, p.ID(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Stock() != 5 {
		t.Errorf("want stock=5, got %d", updated.Stock())
	}
}

func TestDecreaseStock_InsufficientStock(t *testing.T) {
	svc, _ := setup(t)
	ctx := context.Background()

	p, _ := svc.CreateProduct(ctx, "Widget", 500, 2)
	_, err := svc.DecreaseStock(ctx, p.ID(), 99)
	if err == nil {
		t.Fatal("expected error")
	}
	if !domainerrors.Is(err, domainerrors.ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
	}
}
