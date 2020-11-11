package processor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type accentuator struct {
	httpclient *http.Client
	url        string
}

//NewAccentuator creates new processor
func NewAccentuator(urlStr string) (synthesizer.Processor, error) {
	res := &accentuator{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *accentuator) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := mapAccentInput(data)
	if len(inData) > 0 {
		abbrResult, err := p.accentCall(inData)
		if err != nil {
			return err
		}
		err = mapAccentOutput(data, abbrResult)
		if err != nil {
			return err
		}
	} else {
		goapp.Log.Debug("Skip accenter - no data in")
	}
	return nil
}

type accentOutputElement struct {
	Accent []accent `json:"accent"`
	Word   string   `json:"word"`
}

type accent struct {
	MF       string                      `json:"mf"`
	Mi       string                      `json:"mi"`
	MiVdu    string                      `json:"mi_vdu"`
	Mih      string                      `json:"mih"`
	Error    string                      `json:"error"`
	Variants []synthesizer.AccentVariant `json:"variants"`
}

func (p *accentuator) accentCall(data []string) ([]accentOutputElement, error) {
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
		return nil, errors.New("Can't resolve accents")
	}
	var res []accentOutputElement
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func mapAccentInput(data *synthesizer.TTSData) []string {
	res := []string{}
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator == "" && w.UserTranscription == "" {
			res = append(res, w.Tagged.Word)
		}
	}
	return res
}

func mapAccentOutput(data *synthesizer.TTSData, out []accentOutputElement) error {
	i := 0
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator == "" && w.UserTranscription == "" {
			if len(out) <= i {
				return errors.New("Wrong accent result")
			}
			err := setAccent(w, out[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

func setAccent(w *synthesizer.ProcessedWord, out accentOutputElement) error {
	if w.Tagged.Word != out.Word {
		return errors.Errorf("Words do not match '%s' vs '%s'", w.Tagged.Word, out.Word)
	}
	w.AccentVariant = findBestAccentVariant(out.Accent, w.Tagged.Mi)
	return nil
}

func findBestAccentVariant(acc []accent, mi string) *synthesizer.AccentVariant {
	find := func(fa func(a *accent) bool, fv func(v *synthesizer.AccentVariant) bool) *synthesizer.AccentVariant {
		for _, a := range acc {
			if fa(&a) {
				for _, v := range a.Variants {
					if fv(&v) {
						return &v
					}
				}
			}
		}
		return nil
	}
	fIsAccent := func(v *synthesizer.AccentVariant) bool { return v.Accent > 0 }

	if res := find(func(a *accent) bool { return a.Error == "" && a.MiVdu == mi }, fIsAccent); res != nil {
		return res
	}
	// no mi filter
	if res := find(func(a *accent) bool { return a.Error == "" }, fIsAccent); res != nil {
		return res
	}
	//no filter
	return find(func(a *accent) bool { return true }, func(v *synthesizer.AccentVariant) bool { return true })
}
