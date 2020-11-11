package synthesizer

import "github.com/airenas/tts-line/internal/pkg/service/api"

//TTSData working data for one request
type TTSData struct {
	OriginalText       string
	Text               string
	TextWithNumbers    string
	Audio              string
	Words              []*ProcessedWord
	ValidationFailures []api.ValidateFailure
}

//ProcessedWord keeps one word info
type ProcessedWord struct {
	Tagged            TaggedWord
	UserTranscription string
	UserSyllables     string
	TranscriptionWord string
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Separator string
	Word      string
	Mi        string
	Lemma     string
}
