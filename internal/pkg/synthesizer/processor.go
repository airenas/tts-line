package synthesizer

import (
	"github.com/airenas/tts-line/internal/pkg/service/api"
)

//Processor worder interface
type Processor interface {
	Process(*TTSData) error
}

//MainWorker does synthesis work
type MainWorker struct {
	Processors []Processor
}

//Work is main method
func (mw *MainWorker) Work(text string) (*api.Result, error) {
	data := &TTSData{}
	data.OriginalText = text
	err := mw.processAll(data)
	if err != nil {
		return nil, err
	}
	return mapResult(data), nil
}

func (mw *MainWorker) processAll(data *TTSData) error {
	for _, pr := range mw.Processors {
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
		res.AudioAsString = data.TextWithNumbers
	}
	return res
}
