package processor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type mp3Converter struct {
	httpclient *http.Client
	url        string
}

//NewMP3 creates new processor
func NewMP3(urlStr string) (synthesizer.Processor, error) {
	res := &mp3Converter{}
	var err error
	res.url, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *mp3Converter) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	inData := mp3Input{Data: data.Audio}
	resp, err := p.mp3Call(&inData)
	if err != nil {
		return err
	}
	data.AudioMP3 = resp.Data
	return nil
}

type mp3Input struct {
	Data string `json:"audio"`
}

type mp3Output struct {
	Data string `json:"audio"`
}

func (p *mp3Converter) mp3Call(data *mp3Input) (*mp3Output, error) {
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
	var res mp3Output
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
