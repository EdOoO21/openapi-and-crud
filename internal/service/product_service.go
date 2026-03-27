package service

import (
	"context"
	"errors"

	"github.com/EdOoO21/openapi-and-crud/internal/api"
	"github.com/EdOoO21/openapi-and-crud/internal/repository"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(r *repository.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

func (s *ProductService) Create(ctx context.Context, in api.ProductCreate, sellerID uuid.UUID) (*repository.Product, error) {
	p := &repository.Product{
		ID:          uuid.New(),
		Name:        in.Name,
		Description: in.Description,
		Price:       float64(in.Price),
		Stock:       in.Stock,
		Category:    in.Category,
		Status:      "ACTIVE",
		SellerID:    sellerID,
	}
	if in.Status != nil {
		p.Status = string(*in.Status)
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, in api.ProductUpdate) (*repository.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	p.Name = in.Name
	p.Description = in.Description
	p.Price = float64(in.Price)
	p.Stock = in.Stock
	p.Category = in.Category
	p.Status = string(in.Status)

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *ProductService) List(ctx context.Context, page, size int, status *string, category *string) ([]repository.Product, int, error) {
	items, total, err := s.repo.List(ctx, page, size, status, category)
	return items, total, err
}
