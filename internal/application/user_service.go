package application

import (
	"context"
	"errors"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain"
	"fmt"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, user *domain.User) error
	// InviteUserToOrg(ctx context.Context, user *domain.User)
	GetUserByEmail(ctx context.Context, email string, password string) (*domain.User, error)
}

type UserService struct {
	userRepo UserRepository
	projRepo ProjectRepository
}

func NewUserService(userRepo UserRepository, projRepo ProjectRepository) *UserService {
	return &UserService{userRepo: userRepo, projRepo: projRepo}
}

func (s *UserService) UserLogin(ctx context.Context, email string, password string) (*domain.User, error) {
	hashedStr, err := access.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	user, err := s.userRepo.GetUserByEmail(ctx, email, hashedStr)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrUserNotFound
	} else if errors.Is(err, domain.ErrUserAlreadyExists) {
		return nil, domain.ErrUserAlreadyExists
	}
	return user, nil
}

func (s *UserService) CreateUserAccount(ctx context.Context, user *domain.User) (*domain.User, error) {
	hashedStr, err := access.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	user, err = s.userRepo.GetUserByEmail(ctx, user.Email, hashedStr)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrUserAlreadyExists
	}
	user = &domain.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: hashedStr,
	}
	err = s.userRepo.CreateNewUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	// I don't need to get the projects here bc the new user won't have any
	return user, nil
}

func (s *UserService) InviteUserToOrg(ctx context.Context, userID int) error {
	// check that the user doesn't already belong to the org

	// invite the user to the org

	// add user to the "staged" project data
	return nil
}
