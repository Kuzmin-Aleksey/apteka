package failure

import "errors"

type ServerError struct {
	baseError
}

func NewServerError(msg string, status int) error {
	return ServerError{
		baseError: newBaseError(msg, status),
	}
}

func (err ServerError) Error() string {
	return "server error: " + err.baseError.Error()
}

func IsServerError(err error) bool {
	return errors.As(err, new(ServerError))
}
