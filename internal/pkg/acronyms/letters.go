package acronyms

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
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
	cwr := []rune{}
	for _, l := range wr {
		d, ok := letters[l]
		if !ok {
			goapp.Log.Warnf("Unknown letter: '%s'", string(l))
			continue
		}
		if d.newWord {
			if len(cwr) > 0 {
				result = append(result, *makeResultWord(cwr))
			}
			result = append(result, *makeResultWord([]rune{l}))
			cwr = []rune{}
		} else {
			cwr = append(cwr, l)
		}
	}
	if len(cwr) > 0 {
		result = append(result, *makeResultWord(cwr))
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
