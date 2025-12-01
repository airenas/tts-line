package acronyms

import (
	"bufio"
	"io"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms/model"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/transcription"

	"github.com/pkg/errors"
)

// Acronyms keeps acronyms processor
type Acronyms struct {
	d map[string][]*data
}

type data struct {
	word  string
	sylls string
	tr    string
}

// New will instantiate aa acronyms processor
func New(r io.Reader) (*Acronyms, error) {
	scanner := bufio.NewScanner(r)
	acr := make(map[string][]*data)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			strs := strings.Split(line, " ")
			if len(strs) < 2 {
				return nil, errors.New("Can not process line " + line)
			}
			k := strings.ToLower(strings.TrimSpace(strs[0]))
			d := parse(strs[1:])
			acr[k] = d
		}
	}
	if len(acr) == 0 {
		return nil, errors.New("No acronyms loaded")
	}
	goapp.Log.Info().Msgf("Loaded %d acronyms", len(acr))
	return &Acronyms{d: acr}, nil
}

func parse(strs []string) []*data {
	r := make([]*data, 0)
	for _, l := range strs {
		d := transcription.Parse(l)
		if d != nil {
			r = append(r, &data{word: d.Word, sylls: d.Sylls, tr: d.Transcription})
		}
	}
	return r
}

func (s *Acronyms) Process(input *model.Input) ([]api.ResultWord, error) {
	var result []api.ResultWord
	wl := strings.ToLower(input.Word)
	r, ok := s.d[wl]
	if ok {
		for _, d := range r {
			w := input.Word
			if len(r) > 1 {
				w = d.word
			}
			result = append(result, api.ResultWord{Word: w, WordTrans: d.word, Syll: d.sylls, UserTrans: d.tr})
		}
	}
	return result, nil
}
