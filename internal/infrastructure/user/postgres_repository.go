package infrastructure

import (
	"context"
	"errors"
	"filmPackager/internal/domain/user"
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

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string, password string) (*user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE email = $1`
	var existingUser user.User
	err := r.db.QueryRow(ctx, query, email).Scan(&existingUser.Id, &existingUser.Name, &existingUser.Email, &existingUser.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user: %v", err)
	}
	return &existingUser, nil
}

func (r *PostgresUserRepository) CreateNewUser(ctx context.Context, user *user.User) error {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.Password).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}
