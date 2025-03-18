package userservice

import (
	"errors"
)

func verifyFirstAndSecondPasswords(password, secondPassword string) error {
	if password != secondPassword {
		return errors.New("Error: passwords do not match!")
	}
	if password == "" || secondPassword == "" {
		return errors.New("Error: password fields left blank!")
	}
	return nil
}
