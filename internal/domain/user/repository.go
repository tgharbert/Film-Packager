package user

import (
	"context"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, user *User) error
	// InviteUserToOrg(ctx context.Context, user *domain.User)
	GetUserByEmail(ctx context.Context, email string, password string) (*User, error)
}
