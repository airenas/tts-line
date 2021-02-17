package synthesizer

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"

	"github.com/airenas/tts-line/internal/pkg/service/api"
)

//Processor interface
type Processor interface {
	Process(*TTSData) error
}

//MainWorker does synthesis work
type MainWorker struct {
	processors      []Processor
	AllowCustomCode bool
}

//Work is main method
func (mw *MainWorker) Work(input *api.TTSRequestConfig) (*api.Result, error) {
	data := &TTSData{}
	data.OriginalText = input.Text
	data.Input = input
	data.RequestID = uuid.NewString()
	if mw.AllowCustomCode {
		tryCustomCode(data)
	}
	err := mw.processAll(data)
	if err != nil {
		return nil, err
	}
	return mapResult(data), nil
}

//Add adds a processor to the end
func (mw *MainWorker) Add(pr Processor) {
	mw.processors = append(mw.processors, pr)
}

func (mw *MainWorker) processAll(data *TTSData) error {
	for _, pr := range mw.processors {
		err := pr.Process(data)
		if err != nil {
			return err
		}
		if len(data.ValidationFailures) > 0 {
			return nil
		}
	}
	return nil
}

func mapResult(data *TTSData) *api.Result {
	res := &api.Result{}
	if len(data.ValidationFailures) > 0 {
		res.ValidationFailures = data.ValidationFailures
	} else {
		res.AudioAsString = data.AudioMP3
		if data.Input.OutputTextFormat != api.TextNone {
			if data.Input.AllowCollectData {
				res.RequestID = data.RequestID
			}
			if (data.Input.OutputTextFormat == api.TextNormalized) {
				res.Text = data.TextWithNumbers
			}
		}
	}
	return res
}

func tryCustomCode(data *TTSData) {
	if strings.HasPrefix(data.OriginalText, "##AM:") {
		data.OriginalText = data.OriginalText[len("##AM:"):]
		data.Cfg.JustAM = true
		goapp.Log.Infof("Start from AM")
	}
}
