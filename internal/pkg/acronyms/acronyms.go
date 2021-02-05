package acronyms

import (
	"bufio"
	"io"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"

	"github.com/pkg/errors"
)

//Acronyms keeps acronyms processor
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
	goapp.Log.Infof("Loaded %d acronyms", len(acr))
	return &Acronyms{d: acr}, nil
}

func parse(strs []string) []*data {
	r := make([]*data, 0)
	for _, l := range strs {
		d := parseTrans(l)
		if d != nil {
			r = append(r, d)
		}
	}
	return r
}

func parseTrans(str string) *data {
	var r data
	r.tr = strings.ReplaceAll(str, "-", "")
	r.sylls = trimAccent(str)
	r.sylls = strings.ToLower(r.sylls)
	r.word = strings.ReplaceAll(r.sylls, "-", "")
	return &r
}

func trimAccent(s string) string {
	r := strings.ReplaceAll(s, "3", "")
	r = strings.ReplaceAll(r, "4", "")
	r = strings.ReplaceAll(r, "9", "")
	return r
}

// Process returns the next random saying
func (s *Acronyms) Process(word, mi string) ([]api.ResultWord, error) {
	var result []api.ResultWord
	wl := strings.ToLower(word)
	r, ok := s.d[wl]
	if ok {
		for _, d := range r {
			w := word
			if len(r) > 1 {
				w = d.word
			}
			result = append(result, api.ResultWord{Word: w, WordTrans: d.word, Syll: d.sylls, UserTrans: d.tr})
		}
	}
	return result, nil
}
