package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*User, error)
	GetUserByName(ctx context.Context, userName string) (*User, error)
	GetAllUsersByName(ctx context.Context, userName string) ([]User, error)
	GetUsersByIDs(ctx context.Context, userIds []uuid.UUID) ([]User, error)
	GetAllNewUsersByName(ctx context.Context, term string, userIDs []uuid.UUID) ([]User, error)
	UpdateUserByID(ctx context.Context, user *User) error
}
