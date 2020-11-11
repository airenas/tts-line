package synthesizer

import "github.com/airenas/tts-line/internal/pkg/service/api"

//TTSData working data for one request
type TTSData struct {
	OriginalText       string
	Text               string
	TextWithNumbers    string
	Words              []*ProcessedWord
	ValidationFailures []api.ValidateFailure
	Spectogram         string
	Audio              string
}

//ProcessedWord keeps one word info
type ProcessedWord struct {
	Tagged            TaggedWord
	UserTranscription string
	UserSyllables     string
	TranscriptionWord string
	AccentVariant     *AccentVariant
	Transcription     string
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Separator string
	Word      string
	Mi        string
	Lemma     string
}

//AccentVariant - accenters's result
type AccentVariant struct {
	Accent   int     `json:"accent"`
	Accented string  `json:"accented"`
	Ml       string  `json:"ml"`
	Syll     string  `json:"syll"`
	Usage    float64 `json:"usage"`
}
