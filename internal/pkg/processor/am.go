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

type amodel struct {
	httpclient *http.Client
	url        string
}

//NewAcousticModel creates new processor
func NewAcousticModel(urlStr string) (synthesizer.Processor, error) {
	res := &amodel{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *amodel) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := mapAMInput(data)
	resp, err := p.amCall(inData)
	if err != nil {
		return err
	}
	err = mapAMOutput(data, resp)
	if err != nil {
		return err
	}
	return nil
}

type amInput struct {
	Text string `json:"text"`
}

type amOutput struct {
	Data string `json:"data"`
}

func (p *amodel) amCall(data *amInput) (*amOutput, error) {
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
	var res amOutput
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func mapAMInput(data *synthesizer.TTSData) *amInput {
	res := &amInput{}
	var sb strings.Builder
	space := " "
	pause := "<space>"
	sb.WriteString(pause)
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator != "" {
			sep := getSep(tgw.Separator)
			if sep != "" {
				sb.WriteString(space + sep)
			}
			if sentenceEnd(sep) {
				sb.WriteString(space + pause)
			}
		} else {
			phns := strings.Split(w.Transcription, " ")
			for _, p := range phns {
				if !skipPhn(p) {
					sb.WriteString(space + p)
				}
			}
		}
	}
	res.Text = sb.String()
	return res
}

func mapAMOutput(data *synthesizer.TTSData, out *amOutput) error {
	data.Spectogram = out.Data
	return nil
}

func getSep(s string) string {
	for _, sep := range []string{",", ".", "!", "?", "..."} {
		if s == sep {
			return s
		}
	}
	return ""
}

func sentenceEnd(s string) bool {
	for _, sep := range []string{".", "!", "?", "..."} {
		if s == sep {
			return true
		}
	}
	return false
}

func skipPhn(s string) bool {
	return s == "-"
}
