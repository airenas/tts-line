package clitics

import (
	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
)

//Processor processes words
type Processor struct {
	clitics map[string]bool
}

// NewProcessor will instantiate a processor
func NewProcessor(clitics map[string]bool) (*Processor, error) {
	return &Processor{clitics: clitics}, nil
}

// Process process words
func (s *Processor) Process(words []*api.CliticsInput) ([]*api.CliticsOutput, error) {
	res := make([]*api.CliticsOutput, 0)
	for i := 0; i < len(words); i++ {
		w := words[i]
		if w.Type == "WORD" && s.clitics[w.String] {
			ni, ok := nextWord(words[i+1:])
			if ok {
				nw := words[ni+i+1]
				res = append(res, &api.CliticsOutput{ID: w.ID, Type: "CLITIC", AccentType: "NONE", Pos: 0})
				res = append(res, &api.CliticsOutput{ID: nw.ID, Type: "CLITIC", AccentType: "NORMAL", Pos: 1})
				i = ni
			}
		}
	}
	return res, nil
}

func nextWord(words []*api.CliticsInput) (int, bool) {
	for i := 0; i < len(words); i++ {
		w := words[i]
		if w.Type == "WORD" {
			return i, true
		} else if w.Type != "SPACE" {
			return 0, false
		}
	}
	return 0, false
}
