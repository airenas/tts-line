package synthesizer

import (
	"github.com/airenas/tts-line/internal/pkg/service/api"
)

//TTSData working data for one request
type TTSData struct {
	Input              *api.TTSRequestConfig
	RequestID          string
	Cfg                TTSConfig
	OriginalText       string
	PreviousText       string // text of previous request loaded by requestID
	Text               string
	TextWithNumbers    string
	ValidationFailures []api.ValidateFailure

	Words []*ProcessedWord
	Parts []*TTSDataPart

	Audio         string
	AudioMP3      string
	AudioDuration float64 // in secs
}

//TTSConfig some TTS configuration
type TTSConfig struct {
	JustAM bool
	Input  *api.TTSRequestConfig
}

//TTSDataPart partial tts data
type TTSDataPart struct {
	Text       string
	Cfg        *TTSConfig
	First      bool
	Words      []*ProcessedWord
	Spectogram string
	Audio      string
}

//ProcessedWord keeps one word info
type ProcessedWord struct {
	Tagged            TaggedWord
	UserTranscription string
	UserSyllables     string
	TranscriptionWord string
	AccentVariant     *AccentVariant
	UserAccent        int
	Clitic            Clitic
	Transcription     string
}

type CliticAccentEnum int

const (
	CliticsUnused CliticAccentEnum = iota
	CliticsNone
	CliticsCustom
)

type Clitic struct {
	Type   CliticAccentEnum
	Accent int
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Separator   string
	SentenceEnd bool
	Space       bool
	Word        string
	Mi          string
	Lemma       string
}

//AccentVariant - accenters's result
type AccentVariant struct {
	Accent   int     `json:"accent"`
	Accented string  `json:"accented"`
	Ml       string  `json:"ml"`
	Syll     string  `json:"syll"`
	Usage    float64 `json:"usage"`
}

//IsWord returns true if object indicates word
func (tw TaggedWord) IsWord() bool {
	return !tw.SentenceEnd && tw.Separator == "" && !tw.Space
}
