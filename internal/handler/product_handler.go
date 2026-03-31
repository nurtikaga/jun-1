package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nurtikaga/jun-1/internal/app"
	domainerrors "github.com/nurtikaga/jun-1/internal/domain/errors"
	"github.com/nurtikaga/jun-1/internal/domain/product"
)

type ProductHandler struct {
	svc    *app.ProductService
	logger *slog.Logger
}

func NewProductHandler(svc *app.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /products", h.CreateProduct)
	mux.HandleFunc("GET /products/{id}", h.GetProduct)
	mux.HandleFunc("PATCH /products/{id}/price", h.UpdatePrice)
	mux.HandleFunc("PATCH /products/{id}/stock/decrease", h.DecreaseStock)
}

type createProductRequest struct {
	Name         string `json:"name"`
	PriceInCents int64  `json:"price_in_cents"`
	Stock        int    `json:"stock"`
}

type updatePriceRequest struct {
	PriceInCents int64 `json:"price_in_cents"`
}

type decreaseStockRequest struct {
	Quantity int `json:"quantity"`
}

type productResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PriceInCents int64  `json:"price_in_cents"`
	Stock        int    `json:"stock"`
	Status       string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toResponse(p *product.Product) productResponse {
	return productResponse{
		ID:           string(p.ID()),
		Name:         p.Name(),
		PriceInCents: p.Price(),
		Stock:        p.Stock(),
		Status:       string(p.Status()),
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("create product: decode body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.svc.CreateProduct(r.Context(), req.Name, req.PriceInCents, req.Stock)
	if err != nil {
		h.handleDomainError(w, "create product", err)
		return
	}
	h.logger.Info("product created", "product_id", string(p.ID()), "name", p.Name())
	writeJSON(w, http.StatusCreated, toResponse(p))
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := product.ProductID(r.PathValue("id"))
	p, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		h.handleDomainError(w, "get product", err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(p))
}

func (h *ProductHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	id := product.ProductID(r.PathValue("id"))
	var req updatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("update price: decode body", "error", err, "product_id", string(id))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.svc.UpdatePrice(r.Context(), id, req.PriceInCents)
	if err != nil {
		h.handleDomainError(w, "update price", err)
		return
	}
	h.logger.Info("product price updated", "product_id", string(p.ID()), "new_price_cents", p.Price())
	writeJSON(w, http.StatusOK, toResponse(p))
}

func (h *ProductHandler) DecreaseStock(w http.ResponseWriter, r *http.Request) {
	id := product.ProductID(r.PathValue("id"))
	var req decreaseStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("decrease stock: decode body", "error", err, "product_id", string(id))
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.svc.DecreaseStock(r.Context(), id, req.Quantity)
	if err != nil {
		h.handleDomainError(w, "decrease stock", err)
		return
	}
	h.logger.Info("product stock decreased", "product_id", string(p.ID()), "remaining_stock", p.Stock(), "status", string(p.Status()))
	writeJSON(w, http.StatusOK, toResponse(p))
}

func (h *ProductHandler) handleDomainError(w http.ResponseWriter, op string, err error) {
	h.logger.Error(op+" failed", "error", err)
	switch {
	case errors.Is(err, domainerrors.ErrProductNotFound):
		writeError(w, http.StatusNotFound, "product not found")
	case errors.Is(err, domainerrors.ErrInvalidPrice):
		writeError(w, http.StatusUnprocessableEntity, "invalid price")
	case errors.Is(err, domainerrors.ErrInsufficientStock):
		writeError(w, http.StatusUnprocessableEntity, "insufficient stock")
	case errors.Is(err, domainerrors.ErrInvalidQuantity):
		writeError(w, http.StatusUnprocessableEntity, "invalid quantity")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
