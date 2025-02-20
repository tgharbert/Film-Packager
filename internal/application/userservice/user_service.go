package userservice

import (
	"context"
	"errors"

	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo user.UserRepository
	projRepo project.ProjectRepository
}

func NewUserService(userRepo user.UserRepository, projRepo project.ProjectRepository) *UserService {
	return &UserService{userRepo: userRepo, projRepo: projRepo}
}

func (s *UserService) UserLogin(ctx context.Context, email, password string) (*user.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)

	// refactor this ordering
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserNotFound
	}

	fmt.Printf("existing user: '%s' '%s'", existingUser.Password, password)

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if err != nil {
		fmt.Println("error comparing password: ", err)
		return nil, user.ErrInvalidPassword
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

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
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

func (s *UserService) VerifyOldPassword(ctx context.Context, userID uuid.UUID, password string) error {
	existingUser, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if err != nil {
		return user.ErrInvalidPassword
	}

	return nil
}

func (s *UserService) SetNewPassword(ctx context.Context, userID uuid.UUID, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	u, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	hashedStr := string(hash)

	u.Password = hashedStr

	err = s.userRepo.UpdateUserByID(ctx, u)
	if err != nil {
		return fmt.Errorf("error setting new password: %v", err)
	}

	return nil
}
