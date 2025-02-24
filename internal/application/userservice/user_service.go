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
	if email == "" || password == "" {
		return nil, user.ErrMissingLoginField
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)

	// refactor this ordering
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return nil, fmt.Errorf("error checking for existing user: %v", err)
	}
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, user.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if err != nil {
		return nil, user.ErrInvalidPassword
	}

	return existingUser, nil
}

func (s *UserService) CreateUser(ctx context.Context, firstName, lastName, email, password, secondPassword string) (*user.User, error) {
	// see user_utils.go
	err := verifyCreateAccountFields(firstName, lastName, email, password, secondPassword)
	if err != nil {
		return nil, err
	}

	username := fmt.Sprintf("%s %s", firstName, lastName)

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

func (s *UserService) VerifyOldPassword(ctx context.Context, userID uuid.UUID, pw1, pw2 string) error {
	// see user_utils.go
	err := verifyFirstAndSecondPasswords(pw1, pw2)
	if err != nil {
		return err
	}

	existingUser, err := s.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(pw1))
	if err != nil {
		return user.ErrInvalidPassword
	}

	return nil
}

func (s *UserService) SetNewPassword(ctx context.Context, userID uuid.UUID, pw1, pw2 string) error {
	// see user_utils.go
	err := verifyFirstAndSecondPasswords(pw1, pw2)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pw1), bcrypt.DefaultCost)
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
