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
	var cwr []*ldata
	ad := allowDot(wl)
	for _, l := range wr {
		d, ok := letters[string(l)]
		if !ok {
			goapp.Log.Warnf("Unknown letter: '%s'", string(l))
			continue
		}
		if l == '.' && !ad {
			continue
		}
		if d.newWord {
			if len(cwr) > 0 {
				result = append(result, *makeResultWord(cwr))
			}
			result = append(result, *makeResultWord([]*ldata{d}))
			cwr = nil
		} else {
			cwr = append(cwr, d)
		}
	}
	if len(cwr) > 0 {
		result = append(result, *makeResultWord(cwr))
	}
	return result, nil
}

func makeResultWord(wr []*ldata) *api.ResultWord {
	var r api.ResultWord
	ssp := ""
	tr := ""
	lr := ""
	for i, d := range wr {
		if i == len(wr)-1 {
			tr = tr + ssp + d.chAccent
		} else {
			tr = tr + ssp + d.ch
		}
		ssp = "-"
		lr = lr + d.letter
	}
	r.Word = lr
	if r.Word == "." {
		r.Word = "taškas"
	}
	r.UserTrans = strings.ReplaceAll(tr, "-", "")
	r.Syll = trimAccent(strings.ToLower(tr))
	r.WordTrans = strings.ReplaceAll(r.Syll, "-", "")
	return &r
}

func allowDot(w string) bool {
	strs := strings.Split(strings.ToLower(w), ".")
	l := len(strs)
	if l < 2 || (strs[l-1] != "lt" && strs[l-1] != "eu") {
		return false
	}
	return true
}
