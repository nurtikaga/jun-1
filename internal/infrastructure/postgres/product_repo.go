package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
	"github.com/nurtikaga/jun-1/internal/domain/product"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) FindByID(ctx context.Context, id product.ProductID) (*product.Product, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, price, stock, status FROM products WHERE id = $1`,
		string(id),
	)

	var (
		pid    string
		name   string
		price  int64
		stock  int
		status string
	)

	if err := row.Scan(&pid, &name, &price, &stock, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("postgres: find product: %w", err)
	}

	p, err := product.New(product.ProductID(pid), name, price, stock, product.Status(status))
	if err != nil {
		return nil, fmt.Errorf("postgres: reconstruct product: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) Save(ctx context.Context, p *product.Product) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO products (id, name, price, stock, status)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE
		 SET price = EXCLUDED.price,
		     stock = EXCLUDED.stock,
		     status = EXCLUDED.status,
		     updated_at = NOW()`,
		string(p.ID()),
		p.Name(),
		p.Price(),
		p.Stock(),
		string(p.Status()),
	)
	if err != nil {
		return fmt.Errorf("postgres: save product: %w", err)
	}
	return nil
}
