package processor

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

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
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
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

	data.CleanedText = output.Text
	if data.CleanedText == "" {
		return utils.ErrNoInput
	}
	utils.LogData("Output: ", data.CleanedText)
	return nil
}

type normData struct {
	Text string `json:"text"`
}

func (p *cleaner) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
