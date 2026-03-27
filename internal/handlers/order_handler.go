package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EdOoO21/openapi-and-crud/internal/middleware"
	"github.com/EdOoO21/openapi-and-crud/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "TOKEN_INVALID", "unauthorized", nil)
		return
	}
	order, err := h.svc.CreateOrder(r.Context(), uid, req)
	if err != nil {
		switch err {
		case service.ErrOrderLimitExceeded:
			writeAPIError(w, http.StatusTooManyRequests, "ORDER_LIMIT_EXCEEDED", err.Error(), nil)
		case service.ErrOrderHasActive:
			writeAPIError(w, http.StatusConflict, "ORDER_HAS_ACTIVE", err.Error(), nil)
		case service.ErrProductNotFound:
			writeAPIError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", err.Error(), nil)
		case service.ErrProductInactive:
			writeAPIError(w, http.StatusConflict, "PRODUCT_INACTIVE", err.Error(), nil)
		case service.ErrInsufficientStock:
			writeAPIError(w, http.StatusConflict, "INSUFFICIENT_STOCK", err.Error(), nil)
		default:
			writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"order_id": order.ID.String(), "status": order.Status})
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	oid, err := uuid.Parse(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid uuid", nil)
		return
	}
	uid, _ := middleware.GetUserID(r.Context())
	order, items, err := h.svc.GetOrder(r.Context(), uid, oid)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "ORDER_NOT_FOUND", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order": order, "items": items,
	})
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "update order not implemented", nil)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	oid, err := uuid.Parse(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid uuid", nil)
		return
	}
	uid, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "TOKEN_INVALID", "unauthorized", nil)
		return
	}
	err = h.svc.CancelOrder(r.Context(), uid, oid)
	if err != nil {
		if err == service.ErrInvalidStateTransition {
			writeAPIError(w, http.StatusConflict, "INVALID_STATE_TRANSITION", err.Error(), nil)
			return
		}
		if err == service.ErrNotImplemented {
			writeAPIError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "cancel not implemented", nil)
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "CANCELED"})
}
