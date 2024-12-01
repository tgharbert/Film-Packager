package user

import (
	"filmPackager/internal/domain/project"
	"net/mail"
)

type User struct {
	Id          int
	Name        string
	Email       string
	Password    string
	Memberships []project.ProjectOverview
	Invited     []project.ProjectOverview
}

func CreateNewUser(name, email, password string) *User {
	return &User{
		// Id:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: password,
	}
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
