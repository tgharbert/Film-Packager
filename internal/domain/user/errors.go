package user

import "errors"

var ErrUserNotFound = errors.New("user not found!")
var ErrUserAlreadyExists = errors.New("user already exists!")
var ErrInvalidPassword = errors.New("invalid password!")
var ErrMissingLoginField = errors.New("password and email fields must be filled!")
