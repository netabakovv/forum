package errors

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrDeleteFailed      = errors.New("user was not deleted")
)
