package synthesizer

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/pkg/ssml"
)

//Processor interface
type Processor interface {
	Process(*TTSData) error
}

//MainWorker does synthesis work
type MainWorker struct {
	processors      []Processor
	ssmlProcessors  []Processor
	AllowCustomCode bool
}

//Work is main method
func (mw *MainWorker) Work(input *api.TTSRequestConfig) (*api.Result, error) {
	data := &TTSData{}
	data.OriginalText = input.Text
	data.Input = input
	data.Cfg.Input = input
	data.Cfg.Type = SSMLNone
	data.Cfg.Speed = input.Speed
	data.Cfg.Voice = input.Voice
	data.RequestID = input.RequestID
	data.AudioSuffix = input.AudioSuffix
	if input.RequestID == "" {
		data.RequestID = uuid.NewString()
	}
	if mw.AllowCustomCode {
		tryCustomCode(data)
	}
	if len(input.SSMLParts) > 0 {
		data.Cfg.Type = SSMLMain
		var err error
		data.SSMLParts, err = makeSSMLParts(input)
		if err != nil {
			return nil, err
		}
		if err := mw.processAll(mw.ssmlProcessors, data); err != nil {
			return nil, err
		}
	} else {
		if err := mw.processAll(mw.processors, data); err != nil {
			return nil, err
		}
	}
	return mapResult(data)
}

func makeSSMLParts(input *api.TTSRequestConfig) ([]*TTSData, error) {
	var res []*TTSData
	for _, p := range input.SSMLParts {
		switch pc := p.(type) {
		case *ssml.Text:
			data := &TTSData{}
			data.OriginalTextParts = makeTextParts(pc.Texts)
			data.Input = input
			data.Cfg.Input = input
			data.Cfg.Speed = pc.Speed
			data.Cfg.Voice = pc.Voice
			data.Cfg.Type = SSMLText
			data.RequestID = input.RequestID
			if input.RequestID == "" {
				data.RequestID = uuid.NewString()
			}
			res = append(res, data)
		case *ssml.Pause:
			data := &TTSData{}
			data.Cfg.PauseDuration = pc.Duration
			data.Cfg.Type = SSMLPause
			res = append(res, data)
		default:
			return nil, errors.Errorf("unknown SSML part type %T", pc)
		}
	}
	return res, nil
}

func makeTextParts(textPart []ssml.TextPart) []*TTSTextPart {
	res := []*TTSTextPart{}
	for _, tp := range textPart {
		res = append(res, &TTSTextPart{Text: tp.Text, Accented: tp.Accented, Syllables: tp.Syllables, UserOEPal: tp.UserOEPal})
	}
	return res
}

// Add adds a processor to the end
func (mw *MainWorker) Add(pr Processor) {
	mw.processors = append(mw.processors, pr)
}

// AddSSML adds a SSML processor to the end
func (mw *MainWorker) AddSSML(pr Processor) {
	mw.ssmlProcessors = append(mw.ssmlProcessors, pr)
}

func (mw *MainWorker) processAll(processors []Processor, data *TTSData) error {
	for _, pr := range processors {
		err := pr.Process(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func mapResult(data *TTSData) (*api.Result, error) {
	res := &api.Result{}
	res.AudioAsString = data.AudioMP3
	if data.Input.OutputTextFormat != api.TextNone {
		if data.Input.AllowCollectData {
			res.RequestID = data.RequestID
		}
		if data.Input.OutputTextFormat == api.TextNormalized {
			res.Text = strings.Join(data.TextWithNumbers, "")
		} else if data.Input.OutputTextFormat == api.TextAccented {
			var err error
			res.Text, err = mapAccentedText(data)
			if err != nil {
				return nil, err
			}
		} else if data.Input.OutputTextFormat == api.TextNone {
		} else {
			return nil, errors.Errorf("can't process OutputTextFormat %s", data.Input.OutputTextFormat.String())
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

//GetTranscriberAccent return accent from ProcessedWord
func GetTranscriberAccent(w *ProcessedWord) int {
	if w.AccentVariant != nil {
		res := w.AccentVariant.Accent
		if w.UserAccent > 0 {
			res = w.UserAccent
		} else if w.TextPart != nil && w.TextPart.Accented != "" { // was empty accent provided if it not set in w.UserAccent
			res = 0
		} else if w.Clitic.Type == CliticsCustom {
			res = w.Clitic.Accent
		} else if w.Clitic.Type == CliticsNone {
			res = 0
		}
		return res
	}
	return 0
}

// GetProcessorsInfo return info about processors for testing
func (mw *MainWorker) GetProcessorsInfo() string {
	return getInfo(mw.processors)
}

// GetSSMLProcessorsInfo return info about processors for testing
func (mw *MainWorker) GetSSMLProcessorsInfo() string {
	return getInfo(mw.ssmlProcessors)
}

func getInfo(processors []Processor) string {
	res := strings.Builder{}
	nl := ""
	for _, pr := range processors {
		pri, ok := pr.(interface {
			Info() string
		})
		if ok {
			res.WriteString(nl + pri.Info())
			nl = "\n"
		}
	}
	return res.String()
}
