package application

import (
	"context"
	"errors"
	access "filmPackager/internal/auth"

	// "filmPackager/internal/domain"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"fmt"
)

// type UserRepository interface {
// 	CreateNewUser(ctx context.Context, user *user.User) error
// 	// InviteUserToOrg(ctx context.Context, user *user.User)
// 	GetUserByEmail(ctx context.Context, email string, password string) (*user.User, error)
// }

type UserService struct {
	userRepo user.UserRepository
	projRepo project.ProjectRepository
}

func NewUserService(userRepo user.UserRepository, projRepo project.ProjectRepository) *UserService {
	return &UserService{userRepo: userRepo, projRepo: projRepo}
}

func (s *UserService) UserLogin(ctx context.Context, email string, password string) (*user.User, error) {
	hashedStr, err := access.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email, hashedStr)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		fmt.Println("error checking for existing user", err)
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserNotFound
	} else if errors.Is(err, user.ErrUserAlreadyExists) {
		return nil, user.ErrUserAlreadyExists
	}
	return existingUser, nil
}

func (s *UserService) CreateUserAccount(ctx context.Context, newUser *user.User) (*user.User, error) {
	hashedStr, err := access.HashPassword(newUser.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}
	_, err = s.userRepo.GetUserByEmail(ctx, newUser.Email, hashedStr)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserAlreadyExists
	}
	err = s.userRepo.CreateNewUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	// I don't need to get the projects here bc the new user won't have any
	return newUser, nil
}

func (s *UserService) InviteUserToOrg(ctx context.Context, userID int) error {
	// check that the user doesn't already belong to the org

	// invite the user to the org

	// add user to the "staged" project data
	return nil
}
