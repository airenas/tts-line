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

//ErrWordTooLong indicates too long word
type ErrWordTooLong struct {
	Word string
}

//NewErrBadAccent creates new error
func NewErrWordTooLong(word string) *ErrWordTooLong {
	return &ErrWordTooLong{Word: word}
}

func (r *ErrWordTooLong) Error() string {
	return fmt.Sprintf("Wrong accent, too long word: '%s'", r.Word)
}
