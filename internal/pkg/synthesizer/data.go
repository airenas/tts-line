package synthesizer

import "github.com/airenas/tts-line/internal/pkg/service/api"

//TTSData working data for one request
type TTSData struct {
	Cfg                TTSConfig
	OriginalText       string
	Text               string
	TextWithNumbers    string
	ValidationFailures []api.ValidateFailure

	Words []*ProcessedWord
	Parts []*TTSDataPart

	Audio    string
	AudioMP3 string
}

//TTSConfig some TTS configuration
type TTSConfig struct {
	JustAM bool
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
	Transcription     string
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Separator   string
	SentenceEnd bool
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
	return !tw.SentenceEnd && tw.Separator == ""
}
