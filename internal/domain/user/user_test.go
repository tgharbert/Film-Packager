package user_test

import (
	"filmPackager/internal/domain/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	assert := assert.New(t)

	u := user.CreateNewUser("test name", "test email", "test password")
	assert.False(user.IsValidEmail(u.Email))
}
