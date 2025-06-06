package synthesizer

import (
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
)

// TTSData working data for one request
type TTSData struct {
	Input        *api.TTSRequestConfig
	RequestID    string
	Cfg          TTSConfig
	OriginalText string
	PreviousText string // text of previous request loaded by requestID

	CleanedText []string // corresponds to OriginalTextParts

	NormalizedText  []string // text after normalization, array corresponds to OriginalTextParts
	Text            []string // text after cleaning and URL replacement
	TextWithNumbers []string

	AudioSuffix string // add audio suffix if var is set

	Words []*ProcessedWord
	Parts []*TTSDataPart

	Audio           []byte
	AudioMP3        []byte
	AudioLenSeconds float64
	SampleRate      uint32

	OriginalTextParts []*TTSTextPart
	SSMLParts         []*TTSData
}

// TTSTextPart part of the text
type TTSTextPart struct {
	Accented, Text, Syllables, UserOEPal string
}

// TTSConfig some TTS configuration
type TTSConfig struct {
	JustAM bool
	Input  *api.TTSRequestConfig

	Type  SSMLTypeEnum
	Voice string
	Speed float32

	PauseDuration time.Duration
}

// TTSDataPart partial tts data
type TTSDataPart struct {
	Text               string
	Cfg                *TTSConfig
	First              bool
	Words              []*ProcessedWord
	Spectogram         string
	Audio              string
	TranscribedText    string
	TranscribedSymbols []string
	// from AM response
	Durations      []int
	DefaultSilence int
	Step           int
	AudioDurations *AudioDurations
}

type SynthesizedPos struct {
	// in tts steps
	From, To int
}

type AudioDurations struct {
	// in tts steps
	Shift    int
	Duration time.Duration
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
	TextPart          *TTSTextPart
	SynthesizedPos    *SynthesizedPos
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
