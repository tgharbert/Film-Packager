package project

import "errors"

var ErrMemberNotFound = errors.New("member not found")
var ErrMemberAlreadyExists = errors.New("member already exists")
var ErrMemberAlreadyInvited = errors.New("member already invited")
