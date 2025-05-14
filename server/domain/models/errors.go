package models

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var (
	ErrUnknown                = errors.New("unknown error")
	ErrNotFound               = errors.New("not found")
	ErrContextDone            = errors.New("context done")
	ErrInvalidRequest         = errors.New("invalid request")
	ErrDatabaseError          = errors.New("database error")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrInvalidFile            = errors.New("invalid file")
)

type Error struct {
	Msg       string
	DomainErr error
	Initiator string
}

func (e *Error) Error() string {
	return fmt.Sprintf(`%s "%s"  %s`, e.Initiator, e.DomainErr.Error(), e.Msg)
}

func NewError(domainErr error, messages ...any) error {
	err := new(Error)

	for _, msg := range messages {
		if msg == nil {
			continue
		}
		switch v := msg.(type) {
		case error:
			err.Msg += v.Error()
		case string:
			err.Msg += v
		default:
			err.Msg += fmt.Sprint(msg)
		}
		err.Msg += ": "
	}
	err.Msg = strings.TrimRight(err.Msg, ": ")

	err.DomainErr = domainErr

	pc, _, line, _ := runtime.Caller(1)
	err.Initiator = runtime.FuncForPC(pc).Name() + ":" + strconv.Itoa(line)

	return err
}

func GetDomainErr(e error) error {
	var err *Error
	if errors.As(e, &err) {
		return err.DomainErr
	}
	return ErrUnknown
}

func AddError(err error, e error) error {
	if e == nil {
		return err
	}
	var domainErr *Error
	if errors.As(e, &domainErr) {
		domainErr.Msg += ", " + e.Error()
	}
	return errors.New(err.Error() + ", " + e.Error())
}

func ErrorIs(e error, target error) bool {
	domainErr := GetDomainErr(e)

	if domainErr == nil {
		return false
	}

	return errors.Is(domainErr, target)
}
