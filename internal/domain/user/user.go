package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

// this takes in the string of the hashed password
func CreateNewUser(name, email, password string) *User {
	id := uuid.New()
	return &User{
		Id:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
	}
}

// should move to the util package at some point
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
