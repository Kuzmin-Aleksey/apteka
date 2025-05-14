package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	err := NewError(ErrInvalidRequest, "invalid json fields", ErrDatabaseError)
	err = AddError(err, errors.New("added error"))
	fmt.Println(err.Error())
	fmt.Println(GetDomainErr(err))
}

func TestAnything(t *testing.T) {
	//f := true

}
