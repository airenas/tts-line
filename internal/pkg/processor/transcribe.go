package processor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type transcriber struct {
	httpclient *http.Client
	url        string
}

//NewTranscriber creates new processor
func NewTranscriber(urlStr string) (synthesizer.Processor, error) {
	res := &transcriber{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *transcriber) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := mapTransInput(data)
	if len(inData) > 0 {
		resp, err := p.transCall(inData)
		if err != nil {
			return err
		}
		err = mapTransOutput(data, resp)
		if err != nil {
			return err
		}
	} else {
		goapp.Log.Debug("Skip transcriber - no data in")
	}
	return nil
}

type transInput struct {
	Word string `json:"word"`
	Syll string `json:"syll"`
	User string `json:"user"`
	Ml   string `json:"ml"`
	Rc   string `json:"rc"`
	Acc  int    `json:"acc"`
}

type transOutput struct {
	Transcription []trans `json:"transcription"`
	Word          string  `json:"word"`
}

type trans struct {
	Transcription string `json:"transcription"`
}

type aaccent struct {
	MF       string                      `json:"mf"`
	Mi       string                      `json:"mi"`
	MiVdu    string                      `json:"mi_vdu"`
	Mih      string                      `json:"mih"`
	Error    string                      `json:"error"`
	Variants []synthesizer.AccentVariant `json:"variants"`
}

func (p *transcriber) transCall(data []*transInput) ([]transOutput, error) {
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
	var res []transOutput
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func mapTransInput(data *synthesizer.TTSData) []*transInput {
	res := []*transInput{}
	var pr *transInput
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator != "" {
			pr = nil
		} else {
			ti := &transInput{}
			tword := transWord(w)
			ti.Word = tword
			if w.UserTranscription != "" {
				ti.Syll = w.UserSyllables
				ti.User = w.UserTranscription
				ti.Ml = tword
			} else {
				ti.Acc = w.AccentVariant.Accent
				ti.Syll = w.AccentVariant.Syll
				ti.Ml = w.AccentVariant.Ml
			}
			if pr != nil {
				pr.Rc = tword
			}
			res = append(res, ti)
			pr = ti
		}
	}
	return res
}

func transWord(w *synthesizer.ProcessedWord) string {
	if w.TranscriptionWord != "" {
		return w.TranscriptionWord
	}
	return w.Tagged.Word
}

func mapTransOutput(data *synthesizer.TTSData, out []transOutput) error {
	i := 0
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator == "" {
			if len(out) <= i {
				return errors.New("Wrong transcribe result")
			}
			err := setTrans(w, out[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

func setTrans(w *synthesizer.ProcessedWord, out transOutput) error {
	if transWord(w) != out.Word {
		return errors.Errorf("Words do not match '%s' vs '%s'", transWord(w), out.Word)
	}
	for _, t := range out.Transcription {
		if t.Transcription != "" {
			w.Transcription = dropQMarks(t.Transcription)
			return nil
		}
	}
	return nil
}

func dropQMarks(s string) string {
	return strings.ReplaceAll(s, "?", "")
}
