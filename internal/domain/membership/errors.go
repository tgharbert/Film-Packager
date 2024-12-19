package membership

import "errors"

var (
	ErrUserAlreadyMember = errors.New("user is already a member of the project")
)
