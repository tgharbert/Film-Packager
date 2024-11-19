package infrastructure

import (
	"context"
	"errors"
	"filmPackager/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string, password string) (*domain.User, error) {
	query := `SELECT id, name, email, password FROM user WHERE email = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user: %v", err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) CreateNewUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.Password).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}
