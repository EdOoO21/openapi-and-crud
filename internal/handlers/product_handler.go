package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EdOoO21/openapi-and-crud/internal/api"
	"github.com/EdOoO21/openapi-and-crud/internal/middleware"
	"github.com/EdOoO21/openapi-and-crud/internal/service"

	"github.com/google/uuid"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// ListProducts
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request, params api.ListProductsParams) {
	page := 0
	size := 20
	if params.Page != nil {
		page = int(*params.Page)
	}
	if params.Size != nil {
		size = int(*params.Size)
	}
	var status *string
	if params.Status != nil {
		s := string(*params.Status)
		status = &s
	}
	var category *string
	if params.Category != nil {
		category = params.Category
	}

	items, total, err := h.svc.List(r.Context(), page, size, status, category)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	resp := api.ProductListResponse{
		Items:         []api.ProductResponse{},
		Page:          page,
		Size:          size,
		TotalElements: total,
	}

	for _, it := range items {
		rid := toOpenAPIUUID(it.ID)
		// parse created/updated times if needed; here we keep nil to be simple
		resp.Items = append(resp.Items, api.ProductResponse{
			Id:          rid,
			Name:        it.Name,
			Description: it.Description,
			Price:       float32(it.Price),
			Stock:       it.Stock,
			Category:    it.Category,
			Status:      api.ProductStatus(it.Status),
			SellerId:    nil,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateProduct
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body api.ProductCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}

	sellerID, ok := middleware.GetUserID(r.Context())
	role, _ := middleware.GetUserRole(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "TOKEN_INVALID", "unauthorized", nil)
		return
	}
	if role != "SELLER" && role != "ADMIN" {
		writeAPIError(w, http.StatusForbidden, "ACCESS_DENIED", "only sellers can create products", nil)
		return
	}

	p, err := h.svc.Create(r.Context(), body, sellerID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	resp := api.ProductResponse{
		Id:          toOpenAPIUUID(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       float32(p.Price),
		Stock:       p.Stock,
		Category:    p.Category,
		Status:      api.ProductStatus(p.Status),
		SellerId:    toOpenAPIUUID(p.SellerID),
	}
	writeJSON(w, http.StatusCreated, resp)
}

// DeleteProduct
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, id api.IdParam) {
	// id -> uuid
	gid, err := uuid.Parse(id.String())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid uuid", nil)
		return
	}
	if err := h.svc.SoftDelete(r.Context(), gid); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetProductById
func (h *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request, id api.IdParam) {
	gid, err := uuid.Parse(id.String())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid uuid", nil)
		return
	}
	p, err := h.svc.GetByID(r.Context(), gid)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "product not found", nil)
		return
	}
	resp := api.ProductResponse{
		Id:          toOpenAPIUUID(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       float32(p.Price),
		Stock:       p.Stock,
		Category:    p.Category,
		Status:      api.ProductStatus(p.Status),
		SellerId:    nil,
	}
	writeJSON(w, http.StatusOK, resp)
}

// UpdateProduct
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, id api.IdParam) {
	var body api.ProductUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}
	gid, err := uuid.Parse(id.String())
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid uuid", nil)
		return
	}
	p, err := h.svc.Update(r.Context(), gid, body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	resp := api.ProductResponse{
		Id:          toOpenAPIUUID(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       float32(p.Price),
		Stock:       p.Stock,
		Category:    p.Category,
		Status:      api.ProductStatus(p.Status),
		SellerId:    nil,
	}
	writeJSON(w, http.StatusOK, resp)
}
