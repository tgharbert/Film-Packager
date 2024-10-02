package access

import (
	"time"

	"github.com/golang-jwt/jwt"
	// "github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password string // For storing hashed password
	Role     string // Role, like "admin", "user", etc.
}

type Organization struct {
	ID   int
	Name string
}

type Membership struct {
	UserID         int
	OrganizationID int
	AccessTier     string // Access level like "read", "write", "admin"
}

var jwtKey = []byte("fill_in_l8r")

type Claims struct {
	UserID int
	Email string
	Role string
	jwt.StandardClaims
}

func GenerateJWT(userID int, email string, role string) (string, error) {
	expirationTime := time.Now().Add(48 * time.Hour) // valid for 48 hours
	claims := &Claims{
		UserID: userID,
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

func CheckAccess(userID, orgID int, requiredTier string) (bool, error) {
	var membership Membership
	// query db for access...
	// err := db.QueryRow("SELECT access_tier FROM memberships WHERE user_id = ? AND organization_id = ?", userID, orgID).Scan(&membership.AccessTier)
	// if err != nil {
	// 	return false, err
	// }
	accessHierarchy := map[string]int{
		"read": 1,
		"write": 2,
		"admin": 3,
		"owner": 4,
	}
	userAccessLevel := accessHierarchy[membership.AccessTier]
	requiredAccessLevel := accessHierarchy[requiredTier]
	if userAccessLevel >= requiredAccessLevel {
		return true, nil
	}
	return false, nil
}


