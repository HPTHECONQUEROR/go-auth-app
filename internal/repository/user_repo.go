package repository

import (
	"context"
	"go-auth-app/db"
	"go-auth-app/internal/domain"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`
	_, err := db.DB.Exec(ctx, query, user.Name, user.Email, user.Password)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, created_at FROM users WHERE email = $1`
	row := db.DB.QueryRow(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
