package repository

import (
	"context"
	"go-auth-app/internal/domain"
	"go-auth-app/db"
)

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int) (*domain.User, error)
}

// userRepo implements UserRepository
type userRepo struct{}

// NewUserRepository creates a new instance of userRepo
func NewUserRepository() UserRepository {
	return &userRepo{}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`
	_, err := db.DB.Exec(ctx, query, user.Name, user.Email, user.Password)
	return err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	row := db.DB.QueryRow(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE id = $1`
	row := db.DB.QueryRow(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}