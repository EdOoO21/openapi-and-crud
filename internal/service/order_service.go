package service

import (
	"context"
	"errors"
	"time"

	"github.com/EdOoO21/openapi-and-crud/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrOrderLimitExceeded     = errors.New("order limit exceeded")
	ErrOrderHasActive         = errors.New("user has active order")
	ErrProductNotFound        = errors.New("product not found")
	ErrProductInactive        = errors.New("product inactive")
	ErrInsufficientStock      = errors.New("insufficient stock")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrNotImplemented         = errors.New("not implemented")
)

type OrderService struct {
	repo         *repository.OrderRepository
	prodRepo     *repository.ProductRepository
	userRepo     *repository.UserRepository
	db           *sqlx.DB
	limitMinutes int
}

func NewOrderService(or *repository.OrderRepository, pr *repository.ProductRepository, ur *repository.UserRepository, db *sqlx.DB) *OrderService {
	return &OrderService{
		repo:         or,
		prodRepo:     pr,
		userRepo:     ur,
		db:           db,
		limitMinutes: 1,
	}
}

type CreateOrderItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type CreateOrderRequest struct {
	Items     []CreateOrderItem `json:"items"`
	PromoCode *string           `json:"promo_code,omitempty"`
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, req CreateOrderRequest) (*repository.Order, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Rate limit (check last user_operation CREATE_ORDER)
	var lastOpTime time.Time
	_ = tx.GetContext(ctx, &lastOpTime, `SELECT created_at FROM user_operations WHERE user_id=$1 AND operation_type='CREATE_ORDER' ORDER BY created_at DESC LIMIT 1`, userID)
	if !lastOpTime.IsZero() {
		if time.Since(lastOpTime) < time.Duration(s.limitMinutes)*time.Minute {
			return nil, ErrOrderLimitExceeded
		}
	}

	// 2. Active orders
	var cnt int
	_ = tx.GetContext(ctx, &cnt, `SELECT count(1) FROM orders WHERE user_id=$1 AND status IN ('CREATED','PAYMENT_PENDING')`, userID)
	if cnt > 0 {
		return nil, ErrOrderHasActive
	}

	// 3..5. Check products and reserve stock
	type reserved struct {
		Product repository.Product
		Qty     int
		Price   float64
	}
	var reserves []reserved
	for _, it := range req.Items {
		var p repository.Product
		if err := tx.Get(&p, `SELECT id,name,description,price,stock,category,status,seller_id,created_at,updated_at FROM products WHERE id=$1 FOR UPDATE`, it.ProductID); err != nil {
			return nil, ErrProductNotFound
		}
		if p.Status != "ACTIVE" {
			return nil, ErrProductInactive
		}
		if p.Stock < it.Quantity {
			return nil, ErrInsufficientStock
		}
		// reserve
		if _, err := tx.Exec(`UPDATE products SET stock = stock - $1 WHERE id=$2`, it.Quantity, it.ProductID); err != nil {
			return nil, err
		}
		reserves = append(reserves, reserved{Product: p, Qty: it.Quantity, Price: p.Price})
	}

	// 6. snapshot prices and 7. total calc (promo basic: only existence & min amount not implemented in detail)
	var total float64
	for _, rsv := range reserves {
		total += rsv.Price * float64(rsv.Qty)
	}

	order := &repository.Order{
		ID:             uuid.New(),
		UserID:         userID,
		Status:         "CREATED",
		TotalAmount:    total,
		DiscountAmount: 0,
	}

	var items []repository.OrderItem
	for _, rsv := range reserves {
		items = append(items, repository.OrderItem{
			ID:           uuid.New(),
			OrderID:      order.ID,
			ProductID:    rsv.Product.ID,
			Quantity:     rsv.Qty,
			PriceAtOrder: rsv.Price,
		})
	}

	if err := s.repo.CreateOrderTx(tx, order, items); err != nil {
		return nil, err
	}
	if err := s.repo.InsertUserOperationTx(tx, userID, "CREATE_ORDER"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, userID, orderID uuid.UUID) (*repository.Order, []repository.OrderItem, error) {
	o, items, err := s.repo.GetOrder(nil, orderID)
	if err != nil {
		return nil, nil, err
	}
	// права: ADMIN может смотреть любую, USER только свои
	if userID != uuid.Nil && o.UserID != userID {
		// если не владелец — нужно проверить роль, но svc не знает роль — контроля возложена на handler
	}
	return o, items, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, userID, orderID uuid.UUID, items []repository.OrderItem) (*repository.Order, error) {
	return nil, ErrNotImplemented
}

func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	return ErrNotImplemented
}
