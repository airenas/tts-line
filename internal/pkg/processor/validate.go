package processor

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/airenas/tts-line/internal/pkg/service/api"

	"github.com/spf13/viper"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type validator struct {
	httpclient *http.Client
	url        string
	checks     []api.Check
}

//NewValidator creates new processor
func NewValidator(config *viper.Viper) (synthesizer.Processor, error) {
	if config == nil {
		return nil, errors.New("No config 'validator'")
	}
	res := &validator{}
	var err error
	res.url, err = checkURL(config.GetString("url"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.httpclient = &http.Client{}
	res.checks, err = initChecks(goapp.Sub(config, "check"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't init checks")
	}
	return res, nil
}

func initChecks(config *viper.Viper) ([]api.Check, error) {
	if config == nil {
		return nil, errors.New("No config 'check'")
	}
	res := make([]api.Check, 0)
	for _, s := range []string{"min_words", "max_words", "no_numbers", "profanity"} {
		v := config.GetInt(s)
		if v > 0 {
			res = append(res, api.Check{ID: s, Value: v})
		}
	}
	return res, nil
}

func (p *validator) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.TextWithNumbers)
	validationResult, err := p.validateText(p.mapValidatorInput(data))
	if err != nil {
		return err
	}
	data.ValidationFailures = validationResult.List
	return nil
}

type input struct {
	Text   string      `json:"text"`
	Words  words       `json:"words"`
	Checks []api.Check `json:"checks"`
}

type words struct {
	List []word `json:"list"`
}

type word struct {
	Mi    string `json:"mi"`
	Lemma string `json:"lemma"`
	Word  string `json:"word"`
}

type output struct {
	List []api.ValidateFailure `json:"list"`
}

func (p *validator) validateText(data *input) (*output, error) {
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
		return nil, errors.New("Can't validate")
	}
	var res output
	err = decodeJSONAndLog(resp.Body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *validator) mapValidatorInput(data *synthesizer.TTSData) *input {
	res := &input{}
	res.Checks = p.checks
	res.Text = data.TextWithNumbers
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator == "" {
			res.Words.List = append(res.Words.List, word{Word: tgw.Word, Lemma: tgw.Lemma, Mi: tgw.Mi})
		}
	}
	return res
}

func decodeJSONAndLog(body io.ReadCloser, res interface{}) error {
	br, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "Can't read body")
	}
	goapp.Log.Debug(string(br))
	err = json.Unmarshal(br, &res)
	if err != nil {
		return errors.Wrap(err, "Can't decode response")
	}
	return nil
}
