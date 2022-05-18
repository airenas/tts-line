package ssml

import "time"

// Part is an abstracted part of SSML input
type Part interface{}

// Text represents Text directive
type Text struct {
	Text  string
	Speed float32
	Voice string
}

// Pause represents Pause directive
type Pause struct {
	IsBreak  bool
	Duration time.Duration
}

// IsPause checks if part is a Pause
func IsPause(p Part) bool {
	_, res := p.(*Pause)
	return res
}
