//go:generate stringer -type=NEREnum
package synthesizer

import (
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/pkg/ssml"
)

// TTSData working data for one request
type TTSData struct {
	Input        *api.TTSRequestConfig
	RequestID    string
	Cfg          TTSConfig
	OriginalText string
	PreviousText string // text of previous request loaded by requestID

	CleanedText []string // corresponds to OriginalTextParts

	NormalizedText []string // text after normalization, array corresponds to OriginalTextParts
	// Text            []string // text after cleaning and URL replacement
	TextWithNumbers []string // text after number replacement to words

	AudioSuffix string // add audio suffix if var is set

	Words []*ProcessedWord
	Parts []*TTSDataPart

	Audio    *AudioData
	AudioMP3 []byte

	OriginalTextParts []*TTSTextPart
	SSMLParts         []*TTSData
}

type AudioData struct {
	Data          []byte
	SampleRate    uint32
	BitsPerSample uint16
	Duration      time.Duration
}

func (a *AudioData) Seconds() float64 {
	if a == nil {
		return 0
	}
	return a.Duration.Seconds()
}

// TTSTextPart part of the text
type TTSTextPart struct {
	Accented, Text, Syllables, UserOEPal, Language string

	InterpretAs       ssml.InterpretAsType
	InterpretAsDetail ssml.InterpretAsDetailType

	Prosodies []*ssml.Prosody

	PauseAfter time.Duration
}

// TTSConfig some TTS configuration
type TTSConfig struct {
	JustAM bool
	Input  *api.TTSRequestConfig

	Type      SSMLTypeEnum
	Voice     string
	SpeedRate float64

	PauseDuration time.Duration
}

// TTSDataPart partial tts data
type TTSDataPart struct {
	Text               string
	Cfg                *TTSConfig
	First              bool
	Words              []*ProcessedWord
	Spectogram         []byte
	Audio              []byte
	TranscribedText    string
	TranscribedSymbols []string
	// from AM response
	Durations      []int
	DefaultSilence int
	Step           int
	Loudness       float64
	LoudnessGain   float64
}

type SynthesizedPos struct {
	// in tts steps
	From, To   int
	StartIndex int // in base duration list

	Durations     []int     //chars
	VolumeChanges []float64 // in dB
}

type AudioPos struct {
	From int // position in bytes
	To   int
}

// ProcessedWord keeps one word info
type ProcessedWord struct {
	Tagged            TaggedWord
	UserTranscription string
	UserSyllables     string
	TranscriptionWord string
	AccentVariant     *AccentVariant
	UserAccent        int
	Clitic            Clitic
	Transcription     string
	Obscene           bool
	LastEmphasisWord  bool
	TextPart          *TTSTextPart
	SynthesizedPos    *SynthesizedPos
	NERType           NEREnum
	IsLastInPart      bool
	AudioPos          *AudioPos
	FromWord          *TaggedWord
}

func (p *ProcessedWord) Clone() *ProcessedWord {
	res := *p
	return &res
}

// CliticAccentEnum contains types of possible clitics
type CliticAccentEnum int

const (
	//CliticsUnused - clitics does not apply for the word
	CliticsUnused CliticAccentEnum = iota

	//CliticsNone - not a clitic
	CliticsNone

	//CliticsCustom - custom clitic type
	CliticsCustom
)

// NEREnum
type NEREnum int

const (
	//NERRegular - regular word
	NERRegular NEREnum = iota
	//NERSingleLetter - single letter word
	NERSingleLetter
	//NERGreekLetters - greek letters
	NERGreekLetters
	//NERReadableSymbol - readable symbol
	NERReadableSymbol
	//NERReadableAllSymbol - readable symbol (all characters)
	NERReadableAllSymbol
)

// Clitic structure
type Clitic struct {
	Type   CliticAccentEnum
	Accent int
}

// TaggedWord - tagger's result
type TaggedWord struct {
	Separator   string
	SentenceEnd bool
	Space       bool
	Word        string
	Mi          string
	Lemma       string
}

// AccentVariant - accenters's result
type AccentVariant struct {
	Accent   int     `json:"accent"`
	Accented string  `json:"accented"`
	Ml       string  `json:"ml"`
	Syll     string  `json:"syll"`
	Usage    float64 `json:"usage"`
}

// SSMLTypeEnum indicates part type: text, pause
type SSMLTypeEnum int

const (
	// SSMLNone - not ssml part
	SSMLNone SSMLTypeEnum = iota
	// SSMLMain - main part
	SSMLMain
	// SSMLText - text part for synthesis
	SSMLText
	// SSMLPause - <p>, <break> part for synthesis
	SSMLPause
)

// IsWord returns true if object indicates word
func (tw TaggedWord) IsWord() bool {
	return !tw.SentenceEnd && tw.Separator == "" && !tw.Space
}

func (tw TaggedWord) TypeStr() string {
	if tw.SentenceEnd {
		return "SENTENCE_END"
	}
	if tw.Separator != "" {
		return "SEPARATOR"
	}
	if tw.Space {
		return "SPACE"
	}
	return "WORD"
}

func (tw TaggedWord) Str() string {
	if tw.IsWord() {
		return tw.Word
	}
	return tw.Separator
}

func (tp *TTSTextPart) EmphasisID() int {
	if tp == nil {
		return 0
	}
	for _, p := range tp.Prosodies {
		if p.Emphasis == ssml.EmphasisTypeModerate || p.Emphasis == ssml.EmphasisTypeStrong {
			return p.ID
		}
	}
	return 0
}
