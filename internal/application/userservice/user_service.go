package userservice

import (
	"context"
	"errors"
	access "filmPackager/internal/auth"

	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserNotFound
	} else if errors.Is(err, user.ErrUserAlreadyExists) {
		return nil, user.ErrUserAlreadyExists
	}

	return existingUser, nil
}

// change to accept user params then build the user object in the service
func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*user.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	hashedStr := string(hash)

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email, hashedStr)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserAlreadyExists
	}
	if existingUser != nil {
		return nil, user.ErrUserAlreadyExists
	}

	newUser := user.CreateNewUser(username, email, hashedStr)
	err = s.userRepo.CreateNewUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	return newUser, nil
}
