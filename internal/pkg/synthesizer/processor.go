package synthesizer

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/accent"
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
	data.Cfg.Input = input
	data.RequestID = input.RequestID
	if input.RequestID == "" {
		data.RequestID = uuid.NewString()
	}
	if mw.AllowCustomCode {
		tryCustomCode(data)
	}
	err := mw.processAll(data)
	if err != nil {
		return nil, err
	}
	return mapResult(data)
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

func mapResult(data *TTSData) (*api.Result, error) {
	res := &api.Result{}
	if len(data.ValidationFailures) > 0 {
		res.ValidationFailures = data.ValidationFailures
		return res, nil
	}

	res.AudioAsString = data.AudioMP3
	if data.Input.OutputTextFormat != api.TextNone {
		if data.Input.AllowCollectData {
			res.RequestID = data.RequestID
		}
		if data.Input.OutputTextFormat == api.TextNormalized {
			res.Text = data.TextWithNumbers
		} else if data.Input.OutputTextFormat == api.TextAccented {
			var err error
			res.Text, err = mapAccentedText(data)
			if err != nil {
				return nil, err
			}
		} else if data.Input.OutputTextFormat == api.TextNone {
		} else {
			return nil, errors.Errorf("Can't process OutputTextFormat %s", data.Input.OutputTextFormat.String())
		}
	}
	return res, nil
}

func tryCustomCode(data *TTSData) {
	if strings.HasPrefix(data.OriginalText, "##AM:") {
		data.OriginalText = data.OriginalText[len("##AM:"):]
		data.Cfg.JustAM = true
		goapp.Log.Infof("Start from AM")
	}
}

func mapAccentedText(data *TTSData) (string, error) {
	res := strings.Builder{}
	for _, p := range data.Parts {
		for _, w := range p.Words {
			tgw := w.Tagged
			if tgw.Space {
				res.WriteString(" ")
			} else if tgw.Separator != "" {
				res.WriteString(tgw.Separator)
			} else if tgw.IsWord() {
				aw, err := accent.ToAccentString(tgw.Word, GetTranscriberAccent(w))
				if err != nil {
					return "", errors.Wrapf(err, "Can't mark accent for %s", tgw.Word)
				}
				res.WriteString(aw)
			}
		}
	}
	return res.String(), nil
}

func GetTranscriberAccent(w *ProcessedWord) int {
	if w.AccentVariant != nil {
		res := w.AccentVariant.Accent
		if w.UserAccent > 0 {
			res = w.UserAccent
		} else if w.Clitic.Type == CliticsCustom {
			res = w.Clitic.Accent
		} else if w.Clitic.Type == CliticsNone {
			res = 0
		}
		return res
	}
	return 0
}
