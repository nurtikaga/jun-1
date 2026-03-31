package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/nurtikaga/jun-1/internal/app"
	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
	"github.com/nurtikaga/jun-1/internal/domain/product"
	"github.com/nurtikaga/jun-1/internal/handler"
)


type memRepo struct {
	data map[product.ProductID]*product.Product
}

func newMemRepo() *memRepo { return &memRepo{data: make(map[product.ProductID]*product.Product)} }

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

func setupServer(t *testing.T) (*httptest.Server, *app.ProductService) {
	t.Helper()
	repo := newMemRepo()
	svc := app.NewProductService(repo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := handler.NewProductHandler(svc, logger)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux), svc
}

func TestCreateProduct_HTTP_Success(t *testing.T) {
	srv, _ := setupServer(t)
	defer srv.Close()

	body := `{"name":"Widget","price_in_cents":999,"stock":100}`
	resp, err := http.Post(srv.URL+"/products", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["name"] != "Widget" {
		t.Errorf("want name=Widget, got %v", result["name"])
	}
}

func TestCreateProduct_HTTP_InvalidBody(t *testing.T) {
	srv, _ := setupServer(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/products", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestCreateProduct_HTTP_NegativePrice(t *testing.T) {
	srv, _ := setupServer(t)
	defer srv.Close()

	body := `{"name":"Bad","price_in_cents":-1,"stock":10}`
	resp, err := http.Post(srv.URL+"/products", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", resp.StatusCode)
	}
}

func TestGetProduct_HTTP_NotFound(t *testing.T) {
	srv, _ := setupServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/products/nonexistent-id")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}
