package authservice

import (
	"errors"
	"filmPackager/internal/domain/user"
)

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
