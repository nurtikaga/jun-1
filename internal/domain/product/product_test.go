package product_test

import (
	"testing"

	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
	"github.com/nurtikaga/jun-1/internal/domain/product"
)

func newTestProduct(t *testing.T) *product.Product {
	t.Helper()
	p, err := product.New(
		product.NewProductID(),
		"Test Widget",
		1000,
		50,
		product.StatusActive,
	)
	if err != nil {
		t.Fatalf("unexpected error creating product: %v", err)
	}
	return p
}


func TestNew_Valid(t *testing.T) {
	p := newTestProduct(t)
	if p.Price() != 1000 {
		t.Errorf("want price=1000, got %d", p.Price())
	}
	if p.Stock() != 50 {
		t.Errorf("want stock=50, got %d", p.Stock())
	}
	if p.Status() != product.StatusActive {
		t.Errorf("want status=ACTIVE, got %s", p.Status())
	}
}

func TestNew_NegativePrice(t *testing.T) {
	_, err := product.New(product.NewProductID(), "X", -1, 10, product.StatusActive)
	if err == nil {
		t.Fatal("expected error for negative price, got nil")
	}
	if !domainerrors.Is(err, domainerrors.ErrInvalidPrice) {
		t.Errorf("want ErrInvalidPrice, got %v", err)
	}
}

func TestNew_NegativeStock(t *testing.T) {
	_, err := product.New(product.NewProductID(), "X", 0, -5, product.StatusActive)
	if err == nil {
		t.Fatal("expected error for negative stock, got nil")
	}
	if !domainerrors.Is(err, domainerrors.ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
	}
}


func TestUpdatePrice_Success(t *testing.T) {
	p := newTestProduct(t)
	if err := p.UpdatePrice(2000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Price() != 2000 {
		t.Errorf("want 2000, got %d", p.Price())
	}
	if !p.Tracker().Dirty(product.FieldPrice) {
		t.Error("FieldPrice should be dirty after UpdatePrice")
	}
}

func TestUpdatePrice_Zero(t *testing.T) {
	p := newTestProduct(t)
	if err := p.UpdatePrice(0); err != nil {
		t.Errorf("price=0 should be allowed, got: %v", err)
	}
}

func TestUpdatePrice_Negative(t *testing.T) {
	p := newTestProduct(t)
	if err := p.UpdatePrice(-1); err == nil {
		t.Fatal("expected error for negative price")
	} else if !domainerrors.Is(err, domainerrors.ErrInvalidPrice) {
		t.Errorf("want ErrInvalidPrice, got %v", err)
	}
}

func TestUpdatePrice_NoDirtyOnFailure(t *testing.T) {
	p := newTestProduct(t)
	_ = p.UpdatePrice(-1)
	if p.Tracker().Dirty(product.FieldPrice) {
		t.Error("FieldPrice must NOT be dirty when UpdatePrice fails")
	}
}


func TestDecreaseStock_Success(t *testing.T) {
	p := newTestProduct(t)
	if err := p.DecreaseStock(10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Stock() != 40 {
		t.Errorf("want stock=40, got %d", p.Stock())
	}
	if !p.Tracker().Dirty(product.FieldStock) {
		t.Error("FieldStock should be dirty")
	}
}

func TestDecreaseStock_ToZero_SetsOutOfStock(t *testing.T) {
	p := newTestProduct(t)
	if err := p.DecreaseStock(50); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Stock() != 0 {
		t.Errorf("want stock=0, got %d", p.Stock())
	}
	if p.Status() != product.StatusOutOfStock {
		t.Errorf("want OUT_OF_STOCK, got %s", p.Status())
	}
	if !p.Tracker().Dirty(product.FieldStatus) {
		t.Error("FieldStatus should be dirty")
	}
}

func TestDecreaseStock_Insufficient(t *testing.T) {
	p := newTestProduct(t)
	err := p.DecreaseStock(100)
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
	if !domainerrors.Is(err, domainerrors.ErrInsufficientStock) {
		t.Errorf("want ErrInsufficientStock, got %v", err)
	}
	if p.Stock() != 50 {
		t.Errorf("stock must not change on error, got %d", p.Stock())
	}
}

func TestDecreaseStock_ZeroQuantity_IsNoOp(t *testing.T) {
	p := newTestProduct(t)
	if err := p.DecreaseStock(0); err != nil {
		t.Errorf("DecreaseStock(0) should be a no-op, got: %v", err)
	}
	if p.Stock() != 50 {
		t.Errorf("stock must not change, got %d", p.Stock())
	}
	if p.Tracker().Dirty(product.FieldStock) {
		t.Error("FieldStock must NOT be dirty after no-op")
	}
}

func TestDecreaseStock_NegativeQuantity(t *testing.T) {
	p := newTestProduct(t)
	err := p.DecreaseStock(-5)
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
	if !domainerrors.Is(err, domainerrors.ErrInvalidQuantity) {
		t.Errorf("want ErrInvalidQuantity, got %v", err)
	}
}


func TestChangeTracker_MultipleMutations(t *testing.T) {
	p := newTestProduct(t)
	_ = p.UpdatePrice(500)
	_ = p.DecreaseStock(50)

	if !p.Tracker().Dirty(product.FieldPrice) {
		t.Error("FieldPrice should be dirty")
	}
	if !p.Tracker().Dirty(product.FieldStock) {
		t.Error("FieldStock should be dirty")
	}
	if !p.Tracker().Dirty(product.FieldStatus) {
		t.Error("FieldStatus should be dirty")
	}
}

func TestChangeTracker_NoDirtyOnFreshProduct(t *testing.T) {
	p := newTestProduct(t)
	if p.Tracker().Dirty(product.FieldPrice) {
		t.Error("fresh product must not have dirty price")
	}
	if p.Tracker().Dirty(product.FieldStock) {
		t.Error("fresh product must not have dirty stock")
	}
	if p.Tracker().Dirty(product.FieldStatus) {
		t.Error("fresh product must not have dirty status")
	}
}
