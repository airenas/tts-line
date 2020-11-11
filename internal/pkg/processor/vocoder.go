package processor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type vocoder struct {
	httpclient *http.Client
	url        string
}

//NewVocoder creates new processor
func NewVocoder(urlStr string) (synthesizer.Processor, error) {
	res := &vocoder{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *vocoder) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := vocInput{Data: data.Spectogram}
	resp, err := p.vocoderCall(&inData)
	if err != nil {
		return err
	}
	data.Audio = resp.Data
	return nil
}

type vocInput struct {
	Data string `json:"data"`
}

type vocOutput struct {
	Data string `json:"data"`
}

func (p *vocoder) vocoderCall(data *vocInput) (*vocOutput, error) {
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
	var res vocOutput
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
