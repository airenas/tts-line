package acronyms

import (
	"strings"

	"github.com/airenas/tts-line/internal/pkg/acronyms/model"
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
func (s *Processor) Process(input *model.Input) ([]api.ResultWord, error) {
	if input.Mode != api.ModeNone {
		return s.letters.Process(input)
	}

	result, err := s.acronyms.Process(input)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result, nil
	}
	if strings.HasPrefix(input.MI, "X") && !canReadAsLetters(input.Word) {
		return []api.ResultWord{{Word: input.Word}}, nil
	}
	return s.letters.Process(input)
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
