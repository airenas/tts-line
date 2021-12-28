package utils

import (
	"fmt"

	"github.com/pkg/errors"
)

//ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

//ErrTextDoesNotMatch indicates old text does not match a new one
var ErrTextDoesNotMatch = errors.New("wrong text")

//ErrTextTooLong indicates too long input
type ErrTextTooLong struct {
	Max, Input int
}

//NewErrBadAccent creates new error
func NewErrTextTooLong(inp, max int) *ErrTextTooLong {
	return &ErrTextTooLong{Max: max, Input: inp}
}

func (r *ErrTextTooLong) Error() string {
	return fmt.Sprintf("text size too long. Passed %d chars. Max %d", r.Input, r.Max)
}

//ErrNoInput indicates no text input
var ErrNoInput = errors.New("no input")

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
