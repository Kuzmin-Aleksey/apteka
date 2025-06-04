package failure

import (
	"fmt"
	"runtime"
	"strconv"
)

type baseError struct {
	Msg       string
	Initiator string
	Status    int
}

func newBaseError(msg string, status int) baseError {
	pc, _, line, _ := runtime.Caller(2)

	return baseError{
		Msg:       msg,
		Initiator: runtime.FuncForPC(pc).Name() + ":" + strconv.Itoa(line),
		Status:    status,
	}
}

func (e baseError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("%s: %d %s", e.Initiator, e.Status, e.Msg)
	}
	return fmt.Sprintf("%s: %s", e.Initiator, e.Msg)
}
