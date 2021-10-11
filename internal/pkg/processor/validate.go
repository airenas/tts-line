package processor

import (
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"

	"github.com/spf13/viper"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type validator struct {
	httpWrap HTTPInvokerJSON
	checks   []api.Check
}

//NewValidator creates new processor
func NewValidator(config *viper.Viper) (synthesizer.Processor, error) {
	if config == nil {
		return nil, errors.New("No config 'validator'")
	}
	res := &validator{}
	var err error
	res.httpWrap, err = utils.NewHTTPWrap(config.GetString("url"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
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
	// workaround fix env prefix
	config.SetEnvPrefix("validator_check")

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
	if p.skip(data) {
		goapp.Log.Info("Skip validator")
		return nil
	}
	inData := p.mapValidatorInput(data)
	var output valOutput
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}
	data.ValidationFailures = output.List
	return nil
}

type valInput struct {
	Text   string      `json:"text"`
	Words  valWords    `json:"words"`
	Checks []api.Check `json:"checks"`
}

type valWords struct {
	List []valWord `json:"list"`
}

type valWord struct {
	Mi    string `json:"mi"`
	Lemma string `json:"lemma"`
	Word  string `json:"word"`
}

type valOutput struct {
	List []api.ValidateFailure `json:"list"`
}

func (p *validator) mapValidatorInput(data *synthesizer.TTSData) *valInput {
	res := &valInput{}
	res.Checks = p.checks
	res.Text = data.TextWithNumbers
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() {
			res.Words.List = append(res.Words.List, valWord{Word: tgw.Word, Lemma: tgw.Lemma, Mi: tgw.Mi})
		}
	}
	return res
}

func (p *validator) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
