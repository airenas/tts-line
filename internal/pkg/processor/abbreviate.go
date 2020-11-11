package processor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type abbreviator struct {
	httpclient *http.Client
	url        string
}

//NewAbbreviator creates new processor
func NewAbbreviator(urlStr string) (synthesizer.Processor, error) {
	res := &abbreviator{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *abbreviator) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := mapAbbrInput(data)
	if len(inData) > 0 {
		abbrResult, err := p.abbrCall(inData)
		if err != nil {
			return err
		}
		mapAbbrOutput(data, abbrResult)
	} else {
		goapp.Log.Debug("Skip abbreviation - no data in")
	}
	return nil
}

type abbrInput struct {
	Word string `json:"word,omitempty"`
	MI   string `json:"mi,omitempty"`
	ID   string `json:"id,omitempty"`
}

type abbrWordOutput struct {
	ID    string           `json:"id,omitempty"`
	Words []abbrResultWord `json:"words,omitempty"`
}

type abbrResultWord struct {
	Word      string `json:"word,omitempty"`
	WordTrans string `json:"wordTrans,omitempty"`
	Syll      string `json:"syll,omitempty"`
	UserTrans string `json:"userTrans,omitempty"`
}

func (p *abbreviator) abbrCall(data []abbrInput) ([]abbrWordOutput, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return nil, err
	}
	goapp.Log.Debug(b.String())
	req, err := http.NewRequest("POST", p.url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	goapp.Log.Debugf("Sending text to: %s", p.url)
	resp, err := p.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New("Can't resolve abbreviations")
	}
	var res []abbrWordOutput
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func mapAbbrInput(data *synthesizer.TTSData) []abbrInput {
	res := []abbrInput{}
	for i, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator == "" && isAbbr(tgw.Mi, tgw.Lemma) {
			res = append(res, abbrInput{Word: tgw.Word, MI: tgw.Mi, ID: strconv.Itoa(i)})
		}
	}
	return res
}

func isAbbr(mi, lemma string) bool {
	return mi != "" && (strings.HasPrefix(mi, "X") || strings.HasPrefix(mi, "Y") || mayBeAbbr(mi, lemma))
}

func mayBeAbbr(mi, lemma string) bool {
	return len(lemma) >= 2 && strings.HasPrefix(mi, "N") && allUpper(lemma)
}

func allUpper(lemma string) bool {
	for _, r := range lemma {
		if unicode.IsLetter(r) && unicode.IsLower(r) {
			return false
		}
	}
	return len(lemma) > 0
}

func mapAbbrOutput(data *synthesizer.TTSData, abbrOut []abbrWordOutput) error {
	om := make(map[int]abbrWordOutput)
	for _, abbr := range abbrOut {
		iID, err := strconv.Atoi(abbr.ID)
		if err != nil {
			return errors.Wrapf(err, "Can't parse ID %s", abbr.ID)
		}
		om[iID] = abbr
	}
	res := make([]*synthesizer.ProcessedWord, 0)
	for i, w := range data.Words {
		abbr, ok := om[i]
		if ok {
			res = append(res, newWords(abbr.Words, w)...)
		} else {
			res = append(res, w)
		}
	}
	data.Words = res
	return nil
}

func newWords(aw []abbrResultWord, w *synthesizer.ProcessedWord) []*synthesizer.ProcessedWord {
	res := []*synthesizer.ProcessedWord{}
	for i, r := range aw {
		wd := synthesizer.ProcessedWord{}
		wd.Tagged.Word = r.Word
		wd.Tagged.Mi = w.Tagged.Mi
		if i == 0 {
			wd.Tagged.Lemma = w.Tagged.Lemma
		}
		wd.UserTranscription = r.UserTrans
		wd.UserSyllables = r.Syll
		wd.TranscriptionWord = r.WordTrans
		res = append(res, &wd)
	}
	return res
}
