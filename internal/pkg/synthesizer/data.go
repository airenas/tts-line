package synthesizer

import (
	"github.com/airenas/tts-line/internal/pkg/service/api"
)

//TTSData working data for one request
type TTSData struct {
	Input           *api.TTSRequestConfig
	RequestID       string
	Cfg             TTSConfig
	OriginalText    string
	PreviousText    string // text of previous request loaded by requestID
	CleanedText     string
	Text            string // text after cleaning and URL replacement
	TextWithNumbers string

	Words []*ProcessedWord
	Parts []*TTSDataPart

	Audio           string
	AudioMP3        string
	AudioLenSeconds float64
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
	Obscene           bool
}

//CliticAccentEnum contains types of possible clitics
type CliticAccentEnum int

const (
	//CliticsUnused - clitics does not apply for the word
	CliticsUnused CliticAccentEnum = iota

	//CliticsNone - not a clitic
	CliticsNone

	//CliticsCustom - custom clitic type
	CliticsCustom
)

//Clitic structure
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
