package access

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password string // For storing hashed password
	Role     string // Roles - writer, producer, director, cinematographer, production designer
}

type Organization struct {
	ID   int
	Name string
}

type Membership struct {
	UserID         int
	OrganizationID int
	AccessTier     string // Access level like "read", "write", "admin" -- can alter roles & push from 'staging area', "owner" - creator
}

type UserInfo struct {
	Id int
	Name  string
	Email string
	Role  string
}

// Potential issue here in not reading the env first??
var jwtKey = []byte(os.Getenv("DEV_DATABASE_URL"))

type Claims struct {
	UserID int
	Name string
	Email string
	Role string
	jwt.StandardClaims
}

func GenerateJWT(userID int, name string, email string, role string) (string, error) {
	expirationTime := time.Now().Add(48 * time.Hour) // valid for 48 hours
	claims := &Claims{
		UserID: userID,
		Name: name,
		Email: email,
		Role: role,
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

func VerifyToken(tokenString string) error {
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

func GetUserNameFromToken(tokenString string) (*UserInfo, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check token signing method etc. here
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil // Return the key used for signing
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	userInfo := &UserInfo{
		Id: claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
	}
	return userInfo, nil
}

func CheckAccess(role string, orgID int, requiredTier string) (bool, error) {
	if role == "owner" {
		return true, nil
	}
	if role != requiredTier {
		return false, nil
	}
	return false, nil
}


func GetUserDataFromCookie(c *fiber.Ctx) (*UserInfo, error) {
	tokenString := c.Cookies("Authorization")
	if tokenString == "" {
		return nil, fmt.Errorf("no token string on cookie")
	}
	tokenString = tokenString[len("Bearer "):]
	userInfo, err := GetUserNameFromToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("unable to get userInfo from token string")
	}
	return userInfo, nil
}