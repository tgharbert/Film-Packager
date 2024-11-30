package user

import "filmPackager/internal/domain/project"

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
		Name:     name,
		Email:    email,
		Password: password,
	}
}
