package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (*User, error) {
	u := &User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, Role: role}
	_, err := r.db.NamedExecContext(ctx, `INSERT INTO users (id,email,password_hash,role) VALUES (:id,:email,:password_hash,:role)`, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `SELECT id,email,password_hash,role,created_at FROM users WHERE email=$1`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `SELECT id,email,password_hash,role,created_at FROM users WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO refresh_tokens (id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4)`, tokenID, userID, tokenHash, expiresAt)
	return err
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenID uuid.UUID) (*RefreshToken, error) {
	var t RefreshToken
	err := r.db.GetContext(ctx, &t, `SELECT id,user_id,token_hash,expires_at,created_at FROM refresh_tokens WHERE id=$1`, tokenID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE id=$1`, tokenID)
	return err
}
