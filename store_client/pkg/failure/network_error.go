package failure

import "errors"

type NetworkError struct {
	baseError
}

func NewNetworkError(msg string) error {
	return NetworkError{
		baseError: newBaseError(msg, 0),
	}
}

func (err NetworkError) Error() string {
	return "network error: " + err.baseError.Error()
}

func IsNetworkError(err error) bool {
	return errors.As(err, new(NetworkError))
}
