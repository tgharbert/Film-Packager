package auth

import (
	"filmPackager/internal/application/authservice"
	"filmPackager/internal/domain/user"

	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type contextKey int

const userKey contextKey = iota

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %v", err)
	}
	hashedStr := string(hash)
	return hashedStr, nil
}

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	UserID uuid.UUID
	jwt.StandardClaims
}

func New(svc *authservice.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("filmpackager")
		if tokenString == "" {
			return c.Next()
		}

		tokenString = tokenString[len("Bearer "):]

		// get user id from token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			return c.Next()
		}
		if !token.Valid {
			return c.Next()
		}

		u, err := svc.UserRepo.GetUserById(c.Context(), claims.UserID)

		if err != nil {
			return c.Next()
		}

		c.Locals(userKey, u)

		return c.Next()
	}
}

func GetUserFromContext(c *fiber.Ctx) *user.User {
	u, ok := c.Locals(userKey).(*user.User)
	if !ok {
		return nil
	}

	return u
}
