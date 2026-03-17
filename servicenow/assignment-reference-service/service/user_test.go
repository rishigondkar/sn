package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailValidation(t *testing.T) {
	assert.True(t, emailRegex.MatchString("a@b.co"))
	assert.True(t, emailRegex.MatchString("user@example.com"))
	assert.False(t, emailRegex.MatchString("invalid"))
	assert.False(t, emailRegex.MatchString(""))
}

func TestErrTypes(t *testing.T) {
	assert.ErrorIs(t, ErrUserNotFound, ErrUserNotFound)
	assert.ErrorIs(t, ErrInvalidEmail, ErrInvalidEmail)
	assert.ErrorIs(t, ErrUsernameTaken, ErrUsernameTaken)
	assert.ErrorIs(t, ErrEmailTaken, ErrEmailTaken)
}
