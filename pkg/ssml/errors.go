package ssml

import "fmt"

// ErrParse indicates an SSML parse error
type ErrParse struct {
	Pos int64
	Msg string
}

func (e *ErrParse) Error() string {
	return fmt.Sprintf("ssml: %d: %s", e.Pos, e.Msg)
}
