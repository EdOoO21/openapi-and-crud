package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	Category    string    `db:"category"`
	Status      string    `db:"status"`
	SellerID    uuid.UUID `db:"seller_id"`
	CreatedAt   string    `db:"created_at"`
	UpdatedAt   string    `db:"updated_at"`
}

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *Product) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	query := `INSERT INTO products (id, name, description, price, stock, category, status, seller_id)
	          VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Stock, p.Category, p.Status, p.SellerID)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	var p Product
	err := r.db.GetContext(ctx, &p, `SELECT id,name,description,price,stock,category,status,seller_id,created_at,updated_at FROM products WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *Product) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET name=$1, description=$2, price=$3, stock=$4, category=$5, status=$6 WHERE id=$7`,
		p.Name, p.Description, p.Price, p.Stock, p.Category, p.Status, p.ID)
	return err
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET status='ARCHIVED' WHERE id=$1`, id)
	return err
}

// List returns items and total count
func (r *ProductRepository) List(ctx context.Context, page, size int, status *string, category *string) ([]Product, int, error) {
	offset := page * size
	where := []string{"true"}
	args := []interface{}{}
	idx := 1

	if status != nil {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, *status)
		idx++
	}
	if category != nil {
		where = append(where, fmt.Sprintf("category = $%d", idx))
		args = append(args, *category)
		idx++
	}
	whereClause := strings.Join(where, " AND ")

	// SELECT items
	q := fmt.Sprintf(`SELECT id,name,description,price,stock,category,status,seller_id,created_at,updated_at
	                  FROM products WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, idx, idx+1)
	argsWithLimit := append(args, size, offset)
	var items []Product
	if err := r.db.SelectContext(ctx, &items, q, argsWithLimit...); err != nil {
		return nil, 0, err
	}

	// COUNT
	countQ := fmt.Sprintf(`SELECT count(1) FROM products WHERE %s`, whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return items, 0, err
	}
	return items, total, nil
}
