package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Order struct {
	ID             uuid.UUID  `db:"id"`
	UserID         uuid.UUID  `db:"user_id"`
	Status         string     `db:"status"`
	PromoCodeID    *uuid.UUID `db:"promo_code_id"`
	TotalAmount    float64    `db:"total_amount"`
	DiscountAmount float64    `db:"discount_amount"`
	CreatedAt      string     `db:"created_at"`
	UpdatedAt      string     `db:"updated_at"`
}

type OrderItem struct {
	ID           uuid.UUID `db:"id"`
	OrderID      uuid.UUID `db:"order_id"`
	ProductID    uuid.UUID `db:"product_id"`
	Quantity     int       `db:"quantity"`
	PriceAtOrder float64   `db:"price_at_order"`
}

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrderTx creates order and items and returns order id. Expects tx provided.
func (r *OrderRepository) CreateOrderTx(tx *sqlx.Tx, order *Order, items []OrderItem) error {
	_, err := tx.Exec(`INSERT INTO orders (id,user_id,status,promo_code_id,total_amount,discount_amount) VALUES ($1,$2,$3,$4,$5,$6)`,
		order.ID, order.UserID, order.Status, order.PromoCodeID, order.TotalAmount, order.DiscountAmount)
	if err != nil {
		return err
	}
	for _, it := range items {
		_, err := tx.Exec(`INSERT INTO order_items (id,order_id,product_id,quantity,price_at_order) VALUES ($1,$2,$3,$4,$5)`,
			it.ID, it.OrderID, it.ProductID, it.Quantity, it.PriceAtOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) GetOrder(tx *sqlx.Tx, id uuid.UUID) (*Order, []OrderItem, error) {
	var o Order
	var items []OrderItem
	var dbq string
	if tx != nil {
		dbq = "SELECT id,user_id,status,promo_code_id,total_amount,discount_amount,created_at,updated_at FROM orders WHERE id=$1"
		if err := tx.Get(&o, dbq, id); err != nil {
			return nil, nil, err
		}
		if err := tx.Select(&items, `SELECT id,order_id,product_id,quantity,price_at_order FROM order_items WHERE order_id=$1`, id); err != nil {
			return nil, nil, err
		}
		return &o, items, nil
	}
	if err := r.db.GetContext(context.Background(), &o, `SELECT id,user_id,status,promo_code_id,total_amount,discount_amount,created_at,updated_at FROM orders WHERE id=$1`, id); err != nil {
		return nil, nil, err
	}
	if err := r.db.SelectContext(context.Background(), &items, `SELECT id,order_id,product_id,quantity,price_at_order FROM order_items WHERE order_id=$1`, id); err != nil {
		return nil, nil, err
	}
	return &o, items, nil
}

// Helper to insert user_operation
func (r *OrderRepository) InsertUserOperationTx(tx *sqlx.Tx, userID uuid.UUID, opType string) error {
	_, err := tx.Exec(`INSERT INTO user_operations (id,user_id,operation_type) VALUES ($1,$2,$3)`, uuid.New(), userID, opType)
	return err
}
