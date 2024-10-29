package acronyms

import (
	"strings"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
)

// Processor processes word
type Processor struct {
	acronyms service.Worker
	letters  service.Worker
}

// NewProcessor will instantiate a processor
func NewProcessor(acronyms, letters service.Worker) (*Processor, error) {
	return &Processor{acronyms: acronyms, letters: letters}, nil
}

// Process process word
func (s *Processor) Process(word, mi string) ([]api.ResultWord, error) {
	result, err := s.acronyms.Process(word, mi)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result, nil
	}
	if strings.HasPrefix(mi, "X") && !canReadAsLetters(word) {
		return []api.ResultWord{{Word: word}}, nil
	}
	return s.letters.Process(word, mi)
}

func canReadAsLetters(word string) bool {
	max := 4
	if len([]rune(word)) <= max {
		return true
	}
	if !allowDot(word) {
		return false
	}
	strs := strings.Split(word, ".")
	for _, s := range strs {
		if len([]rune(s)) > max {
			return false
		}
	}
	return true
}
