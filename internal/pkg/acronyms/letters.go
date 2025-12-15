package acronyms

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms/model"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/transcription"
)

// Letters processes letter abbreviation
type Letters struct {
	letters map[api.Mode]map[string]*ldata
}

// NewLetters will instantiate a abbreviation processor
func NewLetters() (*Letters, error) {
	return &Letters{
		letters: initLetters(),
	}, nil
}

// Process returns the next random saying
func (s *Letters) Process(input *model.Input) ([]api.ResultWord, error) {
	mode := input.Mode
	if mode == api.ModeNone {
		mode = api.ModeCharactersAsWord
	}

	word := input.Word
	var result []api.ResultWord
	wl := strings.ToLower(word)
	if mode != api.ModeAllAsCharacters {
		wl = strings.TrimRight(wl, ".")
	}
	wr := []rune(wl)
	var cwr []*ldata
	ad := allowDot(wl) || mode == api.ModeCharacters || mode == api.ModeAllAsCharacters
	wLen := len(wr)
	var step int
	for i := 0; i < wLen; i = i + step {
		var d *ldata
		ok := false
		if i < wLen-1 && mode == api.ModeCharactersAsWord {
			d, ok = s.getLetters(string(wr[i:i+2]), mode)
		}
		if !ok {
			d, ok = s.getLetters(string(wr[i]), mode)
			step = 1
		} else {
			step = 2
		}
		if !ok {
			goapp.Log.Warn().Str("letter", string(wr[i])).Msg("Unknown")
			continue
		}
		if wr[i] == '.' && !ad {
			continue
		}
		if d.newWord {
			if len(cwr) > 0 {
				result = append(result, *makeResultWord(cwr))
			}
			ldt := d
			for ldt != nil {
				result = append(result, *makeResultWord([]*ldata{ldt}))
				ldt = ldt.next
			}
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

func (s *Letters) getLetters(chars string, mode api.Mode) (*ldata, bool) {
	dict, ok := s.letters[mode]
	if !ok {
		return nil, false
	}
	res, ok := dict[chars]
	return res, ok
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
	r.UserTrans = strings.ReplaceAll(tr, "-", "")
	r.Syll = transcription.TrimAccent(strings.ToLower(tr))
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
