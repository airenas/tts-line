//go:generate stringer -type=InterpretAsType,InterpretAsDetailType -linecomment=true
package ssml

import "time"

// Part is an abstracted part of SSML input
type Part interface{}

// Text represents Text directive
type Text struct {
	Voice     string
	Texts     []TextPart
	Prosodies []*Prosody
}

const (
	MinVolumeChange = -1000.0
)

type PitchChangeKind int

const (
	PitchChangeNone PitchChangeKind = iota
	PitchChangeHertz
	PitchChangeMultiplier
	PitchChangeSemitone
)

type PitchChange struct {
	Kind  PitchChangeKind
	Value float64
}

type EmphasisType int

const (
	EmphasisTypeUnset EmphasisType = iota
	EmphasisTypeReduced
	EmphasisTypeNone
	EmphasisTypeModerate
	EmphasisTypeStrong
)

type InterpretAsType int

const (
	InterpretAsTypeUnset      InterpretAsType = iota // unset
	InterpretAsTypeCharacters                        // characters
)

type InterpretAsDetailType int

const (
	InterpretAsDetailTypeUnset       InterpretAsDetailType = iota // unset
	InterpretAsDetailTypeReadSymbols                              // read-symbols
)

type Prosody struct {
	Rate     float64
	Volume   float64 // in dB, if silent then -1000
	Pitch    PitchChange
	Emphasis EmphasisType
	ID       int
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

// TextPart represents some part of text
type TextPart struct {
	Text              string
	Language          string
	Accented          string
	Syllables         string
	UserOEPal         string // long/short OE and palatalization model
	InterpretAs       InterpretAsType
	InterpretAsDetail InterpretAsDetailType
}
