package utils

import (
	"fmt"

	"github.com/pkg/errors"
)

//ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

//ErrTextDoesNotMatch indicates old text does not match a new one
var ErrTextDoesNotMatch = errors.New("wrong text")

//ErrBadAccent indicate bad accent error
type ErrBadAccent struct {
	BadAccents []string
}

//NewErrBadAccent creates new error
func NewErrBadAccent(badAccents []string) *ErrBadAccent {
	return &ErrBadAccent{BadAccents: badAccents}
}

func (r *ErrBadAccent) Error() string {
	return fmt.Sprintf("Wrong accents: %v", r.BadAccents)
}
