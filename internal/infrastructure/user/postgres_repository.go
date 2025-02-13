package infrastructure

import (
	"context"
	"errors"
	"filmPackager/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
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

func (r *PostgresUserRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE id = $1`
	var existingUser user.User
	err := r.db.QueryRow(ctx, query, userId).Scan(&existingUser.Id, &existingUser.Name, &existingUser.Email, &existingUser.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user: %v", err)
	}
	return &existingUser, nil
}

func (r *PostgresUserRepository) CreateNewUser(ctx context.Context, user *user.User) error {
	query := `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Id, user.Name, user.Email, user.Password).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetAllUsersByName(ctx context.Context, userName string) ([]user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE name = $1`
	var users []user.User
	rows, err := r.db.Query(ctx, query, userName)
	if err != nil {
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var u user.User
		err = rows.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *PostgresUserRepository) GetUserByName(ctx context.Context, userName string) (*user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE name = $1`
	var existingUser user.User
	err := r.db.QueryRow(ctx, query, userName).Scan(&existingUser.Id, &existingUser.Name, &existingUser.Email, &existingUser.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("error scanning user: %v", err)
	}
	return &existingUser, nil
}

func (r *PostgresUserRepository) GetUsersByIDs(ctx context.Context, userIds []uuid.UUID) ([]user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE id = ANY($1)`
	var users []user.User
	rows, err := r.db.Query(ctx, query, userIds)
	if err != nil {
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var u user.User
		err = rows.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *PostgresUserRepository) GetAllNewUsersByTerm(ctx context.Context, term string, userIDs []uuid.UUID) ([]user.User, error) {
	query := `SELECT id, name, email, password FROM users WHERE name ILIKE '%' || $1 || '%' AND id != ALL($2)`
	var users []user.User
	rows, err := r.db.Query(ctx, query, term, userIDs)
	if err != nil {
		return nil, fmt.Errorf("error querying db: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var u user.User
		err = rows.Scan(&u.Id, &u.Name, &u.Email, &u.Password)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		users = append(users, u)
	}
	return users, nil
}
