package clitics

import (
	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
)

//Processor processes words
type Processor struct {
	clitics map[string]bool
	phrases *Phrases
}

// NewProcessor will instantiate a processor
func NewProcessor(clitics map[string]bool, phrases *Phrases) (*Processor, error) {
	return &Processor{clitics: clitics, phrases: phrases}, nil
}

// Process process words
func (s *Processor) Process(words []*api.CliticsInput) ([]*api.CliticsOutput, error) {
	res := make([]*api.CliticsOutput, 0)
	for i := 0; i < len(words); i++ {
		w := words[i]
		added := false
		if w.Type == "WORD" {
			phr, ok := s.phrases.wordMap[w.String]
			phrl, okl := s.phrases.wordMap[w.Lemma]
			if okl {
				phr = append(phr, phrl...)
			}
			if ok || okl {
				for _, ph := range phr {
					nis, ok := isPhrase(words, i, ph)
					if ok {
						for i1, ni := range nis {
							nw := words[ni]
							res = append(res, &api.CliticsOutput{ID: nw.ID, Type: "PHRASE", AccentType: accentType(ph[i1]),
								Accent: accent(ph[i1]), Pos: i1})
							i = ni
						}
						added = true
						break
					}
				}
			}

			if !added && s.clitics[w.String] {
				ni, ok := nextWord(words, i+1)
				if ok {
					res = append(res, &api.CliticsOutput{ID: w.ID, Type: "CLITIC", AccentType: "NONE", Pos: 0})
					nw := words[ni]
					res = append(res, &api.CliticsOutput{ID: nw.ID, Type: "CLITIC", AccentType: "NORMAL", Pos: 1})
					i = ni
				}
			}
		}
	}
	return res, nil
}

func nextWord(words []*api.CliticsInput, from int) (int, bool) {
	for i := from; i < len(words); i++ {
		w := words[i]
		if w.Type == "WORD" {
			return i, true
		} else if w.Type != "SPACE" {
			return 0, false
		}
	}
	return 0, false
}

func isPhrase(words []*api.CliticsInput, from int, ph []*phrase) ([]int, bool) {
	res := make([]int, len(ph))
	wf := from
	for i, pw := range ph {
		wi, ok := nextWord(words, wf)
		if ok {
			wf = wi + 1
			nw := words[wi]
			if pw.isLemma && pw.word == nw.Lemma || !pw.isLemma && pw.word == nw.String {
				res[i] = wi
				continue
			} else {
				return nil, false
			}
		}
		return nil, false
	}
	return res, true
}

func accentType(ph *phrase) string {
	if ph.accent > 0 && !ph.isLemma {
		return api.TypeStatic
	}
	if ph.accent > 0 && ph.isLemma {
		return api.TypeNormal
	}
	return api.TypeNone
}

func accent(ph *phrase) int {
	if ph.accent > 0 && !ph.isLemma {
		return ph.accent
	}
	return 0
}
