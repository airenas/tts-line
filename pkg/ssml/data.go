package ssml

import "time"

// Part is an abstracted part of SSML input
type Part interface{}

// Text represents Text directive
type Text struct {
	Speed        float64
	Voice        string
	Texts        []TextPart
	VolumeChange float64 // in dB, if silent then -1000
}

const (
	MinVolumeChange = -1000.0
)

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

// TextPart represents some part of text
type TextPart struct {
	Text      string
	Accented  string
	Syllables string
	UserOEPal string // long/short OE and palatalization model
}
