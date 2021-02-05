package acronyms

import (
	"strings"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/pkg/errors"
)

//Letters processes letter abbreviation
type Letters struct {
}

// NewLetters will instantiate a abbreviation processor
func NewLetters() (*Letters, error) {
	return &Letters{}, nil
}

// Process returns the next random saying
func (s *Letters) Process(word, mi string) ([]api.ResultWord, error) {
	var result []api.ResultWord
	wl := strings.ToLower(word)
	wl = strings.TrimRight(wl, ".")
	wr := []rune(wl)
	from := 0
	to := 0
	for i, l := range wr {
		d, ok := letters[l]
		if !ok {
			return nil, errors.New("Unknown letter: " + string(l))
		}
		if d.newWord {
			if to > from {
				result = append(result, *makeResultWord(wr[from:to]))
			}
			result = append(result, *makeResultWord(wr[i : i+1]))
			from = i + 1
			to = i + 1
		} else {
			to++
		}
	}
	if to > from {
		result = append(result, *makeResultWord(wr[from:to]))
	}
	return result, nil
}

func makeResultWord(lr []rune) *api.ResultWord {
	var r api.ResultWord
	ssp := ""
	tr := ""
	for i, l := range lr {
		d, _ := letters[l]
		if i == len(lr)-1 {
			tr = tr + ssp + d.chAccent
		} else {
			tr = tr + ssp + d.ch
		}
		ssp = "-"
	}
	r.Word = string(lr)
	r.UserTrans = strings.ReplaceAll(tr, "-", "")
	r.Syll = trimAccent(strings.ToLower(tr))
	r.WordTrans = strings.ReplaceAll(r.Syll, "-", "")
	return &r
}
