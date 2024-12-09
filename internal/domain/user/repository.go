package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string, password string) (*User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*User, error)
}
