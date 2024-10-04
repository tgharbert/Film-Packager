package access

import (
	"time"

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

// // MIDDLEWARE FOR AUTH???
// func RequireAccess(requiredTier string, orgID int, next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// Extract the user information from the JWT token or session
// 			userID := getUserIDFromContext(r)

// 			// Check if the user has the required access
// 			hasAccess, err := CheckAccess(userID, orgID, requiredTier)
// 			if err != nil || !hasAccess {
// 					http.Error(w, "Forbidden", http.StatusForbidden)
// 					return
// 			}

// 			// If access is granted, proceed to the next handler
// 			next.ServeHTTP(w, r)
// 	})
// }

// // Dummy function to extract user ID (assuming it's stored in the context)
// func getUserIDFromContext(r *http.Request) int {
// 	// Extract the user ID from the request's context (set in authentication middleware)
// 	return 123 // Replace this with actual logic to extract the user ID
// }
