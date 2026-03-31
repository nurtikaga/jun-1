package product

import (
	"github.com/google/uuid"

	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
)

type Status string

const (
	StatusActive     Status = "ACTIVE"
	StatusInactive   Status = "INACTIVE"
	StatusOutOfStock Status = "OUT_OF_STOCK"
)

type ProductID string

func NewProductID() ProductID {
	return ProductID(uuid.New().String())
}

type Field int

const (
	FieldPrice Field = iota
	FieldStock
	FieldStatus
)

type ChangeTracker struct {
	dirty map[Field]interface{}
}

func (c *ChangeTracker) Track(field Field, value interface{}) {
	if c.dirty == nil {
		c.dirty = make(map[Field]interface{})
	}
	c.dirty[field] = value
}

func (c *ChangeTracker) Dirty(field Field) bool {
	_, ok := c.dirty[field]
	return ok
}

func (c *ChangeTracker) Changes() map[Field]interface{} {
	out := make(map[Field]interface{}, len(c.dirty))
	for k, v := range c.dirty {
		out[k] = v
	}
	return out
}

type Product struct {
	id      ProductID
	name    string
	price   int64
	stock   int
	status  Status
	tracker ChangeTracker
}

func New(id ProductID, name string, priceInCents int64, stock int, status Status) (*Product, error) {
	if priceInCents < 0 {
		return nil, domainerrors.New("INVALID_PRICE", "initial price cannot be negative", domainerrors.ErrInvalidPrice)
	}
	if stock < 0 {
		return nil, domainerrors.New("INVALID_STOCK", "initial stock cannot be negative", domainerrors.ErrInsufficientStock)
	}
	return &Product{
		id:     id,
		name:   name,
		price:  priceInCents,
		stock:  stock,
		status: status,
	}, nil
}

func (p *Product) ID() ProductID      { return p.id }
func (p *Product) Name() string       { return p.name }
func (p *Product) Price() int64       { return p.price }
func (p *Product) Stock() int         { return p.stock }
func (p *Product) Status() Status     { return p.status }
func (p *Product) Tracker() *ChangeTracker { return &p.tracker }

func (p *Product) UpdatePrice(newPrice int64) error {
	if newPrice < 0 {
		return domainerrors.New("INVALID_PRICE", "price cannot be negative", domainerrors.ErrInvalidPrice)
	}
	p.price = newPrice
	p.tracker.Track(FieldPrice, newPrice)
	return nil
}

func (p *Product) DecreaseStock(quantity int) error {
	if quantity < 0 {
		return domainerrors.New("INVALID_QUANTITY", "quantity must not be negative", domainerrors.ErrInvalidQuantity)
	}
	if quantity == 0 {
		return nil
	}
	if p.stock < quantity {
		return domainerrors.New("INSUFFICIENT_STOCK", "not enough stock to fulfil request", domainerrors.ErrInsufficientStock)
	}
	p.stock -= quantity
	p.tracker.Track(FieldStock, p.stock)
	if p.stock == 0 {
		p.status = StatusOutOfStock
		p.tracker.Track(FieldStatus, p.status)
	}
	return nil
}
