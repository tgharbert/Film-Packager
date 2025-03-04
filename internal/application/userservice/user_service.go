package userservice

import (
	"context"

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
