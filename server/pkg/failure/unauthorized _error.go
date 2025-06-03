package failure

import (
	"errors"
)

type UnauthorizedError struct {
	baseError
}

func NewUnauthorized(msg string) error {
	return UnauthorizedError{
		baseError: newBaseError(msg),
	}
}

func (err UnauthorizedError) Error() string {
	return "unauthorized: " + err.baseError.Error()
}

func IsUnauthorizedError(err error) bool {
	return errors.As(err, new(UnauthorizedError))
}
