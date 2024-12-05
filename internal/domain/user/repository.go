package user

import (
	"context"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string, password string) (*User, error)
	GetUserById(ctx context.Context, userId int) (*User, error)
}
