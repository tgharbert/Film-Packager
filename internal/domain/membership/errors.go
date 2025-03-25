package membership

import "errors"

var (
	ErrUserAlreadyMember  = errors.New("user is already a member of the project")
	ErrSearchTermTooShort = errors.New("please enter at least 3 characters")
)
