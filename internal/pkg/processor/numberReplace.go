package processor

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type numberReplace struct {
	httpclient *http.Client
	url        string
}

//NewNumberReplace creates new processor
func NewNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &numberReplace{}
	var err error
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return nil, errors.New("Can't parse url " + urlStr)
	}
	res.url = urlRes.String()
	res.httpclient = &http.Client{}
	return res, nil
}

func (p *numberReplace) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.Text)
	var err error
	data.TextWithNumbers, err = p.replace(data.Text)
	if err != nil {
		return err
	}
	goapp.Log.Debugf("Out: '%s'", data.TextWithNumbers)
	return nil
}

//Get return duration by calling the service
func (p *numberReplace) replace(text string) (string, error) {
	req, err := http.NewRequest("POST", p.url, strings.NewReader(text))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain")

	goapp.Log.Debugf("Sending text to: %s", p.url)
	resp, err := p.httpclient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", errors.New("Can't get numbers")
	}
	var res string
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", errors.Wrap(err, "Can't decode response")
	}
	return res, nil
}
