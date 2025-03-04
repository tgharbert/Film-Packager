package authservice

import (
	"context"
	"errors"
	"filmPackager/internal/domain/user"
	"time"

	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo user.UserRepository
}

func NewAuthService(userRepo user.UserRepository) *AuthService {
	return &AuthService{UserRepo: userRepo}
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %v", err)
	}
	hashedStr := string(hash)
	return hashedStr, nil
}

// Potential issue here in not reading the env first??
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	UserID uuid.UUID
	jwt.StandardClaims
}

func (s *AuthService) CreateLoginToken(ctx context.Context, userID uuid.UUID, email, password string) (string, error) {
	// get the user info from the token
	u, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error getting user by id: %v", err)
	}
	if u.Password != password {
		return "", fmt.Errorf("invalid password")
	}

	token, err := GenerateJWT(u.Id)
	if err != nil {
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	return token, nil
}

func GenerateJWT(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(48 * time.Hour) // valid for 48 hours
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *AuthService) VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (s *AuthService) CreateNewUser(ctx context.Context, firstName, lastName, email, password, secondPassword string) (string, error) {
	err := verifyCreateAccountFields(firstName, lastName, email, password, secondPassword)
	if err != nil {
		return "", err
	}

	username := fmt.Sprintf("%s %s", firstName, lastName)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %v", err)
	}

	hashedStr := string(hash)

	existingUser, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return "", user.ErrUserAlreadyExists
	}
	if existingUser != nil {
		return "", user.ErrUserAlreadyExists
	}

	newUser := user.CreateNewUser(username, email, hashedStr)
	err = s.UserRepo.CreateNewUser(ctx, newUser)
	if err != nil {
		return "", fmt.Errorf("error creating user: %v", err)
	}

	tokenString, err := GenerateJWT(newUser.Id)
	if err != nil {
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	return tokenString, nil
}

func verifyCreateAccountFields(firstName, lastName, email, password, secondPassword string) error {
	if firstName == "" || lastName == "" {
		return errors.New("Error: please enter first and last name!")
	}
	if email == "" {
		return errors.New("Error: email field left blank!")
	}
	if password != secondPassword {
		return errors.New("Error: passwords do not match!")
	}
	if len(password) < 6 || len(secondPassword) < 6 {
		return errors.New("Error: password needs to be at least 6 characters!")
	}
	if !user.IsValidEmail(email) {
		return errors.New("Error: invalid email address")
	}
	return nil
}
