package processor

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type cleaner struct {
	httpWrap HTTPInvokerJSON
}

//NewCleaner creates new text clean processor
func NewCleaner(urlStr string) (synthesizer.Processor, error) {
	res := &cleaner{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *cleaner) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip clean")
		return nil
	}
	defer goapp.Estimate("Clean")()
	utils.LogData("Input: ", data.OriginalText)
	inData := &normData{Text: data.OriginalText}
	var output normData
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}

	data.Text = output.Text
	if data.Text == "" {
		data.ValidationFailures = []api.ValidateFailure{{Check: api.Check{ID: "no_text"}}}
	}
	utils.LogData("Output: ", data.Text)
	return nil
}

type normData struct {
	Text string `json:"text"`
}

func (p *cleaner) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
