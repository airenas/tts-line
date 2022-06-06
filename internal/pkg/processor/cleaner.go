package processor

import (
	"fmt"
	"strings"
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
	txt := getNormText(data)
	utils.LogData("Input: ", txt)
	inData := &normData{Text: txt}
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

func getNormText(data *synthesizer.TTSData) string {
	if len(data.OriginalTextParts) > 0 {
		res := strings.Builder{}
		for _, s := range data.OriginalTextParts {
			if res.Len() > 0 {
				res.WriteString(" ")
			}
			if s.Accented != "" {
				res.WriteString(s.Accented)
			} else {
				res.WriteString(s.Text)
			}
		}
		return res.String()

	}
	return data.OriginalText
}

type normData struct {
	Text string `json:"text"`
}

func (p *cleaner) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *cleaner) Info() string {
	return fmt.Sprintf("cleaner(%s)", utils.RetrieveInfo(p.httpWrap))
}
