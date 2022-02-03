package processor

import (
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type transcriber struct {
	httpWrap HTTPInvokerJSON
}

//NewTranscriber creates new processor
func NewTranscriber(urlStr string) (synthesizer.PartProcessor, error) {
	res := &transcriber{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *transcriber) Process(data *synthesizer.TTSDataPart) error {
	if p.skip(data) {
		goapp.Log.Info("Skip accentuator")
		return nil
	}

	inData, err := mapTransInput(data)
	if err != nil {
		return err
	}
	if len(inData) > 0 {
		var output []transOutput
		err := p.httpWrap.InvokeJSON(inData, &output)
		if err != nil {
			return err
		}
		err = mapTransOutput(data, output)
		if err != nil {
			return err
		}
	} else {
		goapp.Log.Debug("Skip transcriber - no data in")
	}
	return nil
}

type transInput struct {
	Word string `json:"word"`
	Syll string `json:"syll"`
	User string `json:"user"`
	Ml   string `json:"ml"`
	Rc   string `json:"rc"`
	Acc  int    `json:"acc"`
}

type transOutput struct {
	Transcription []trans `json:"transcription"`
	Word          string  `json:"word"`
	Error         string  `json:"error"`
}

type trans struct {
	Transcription string `json:"transcription"`
}

func mapTransInput(data *synthesizer.TTSDataPart) ([]*transInput, error) {
	res := []*transInput{}
	var pr *transInput
	for _, w := range data.Words {
		tgw := w.Tagged
		if !tgw.IsWord() {
			pr = nil
		} else {
			ti := &transInput{}
			tword := transWord(w)
			ti.Word = tword
			if w.UserTranscription != "" {
				ti.Syll = w.UserSyllables
				ti.User = w.UserTranscription
				ti.Ml = tword
			} else {
				if w.AccentVariant == nil {
					return nil, errors.New("no accent variant for " + tword)
				}
				ti.Acc = synthesizer.GetTranscriberAccent(w)
				ti.Syll = w.AccentVariant.Syll
				ti.Ml = w.AccentVariant.Ml
			}
			if pr != nil {
				pr.Rc = tword
			}
			res = append(res, ti)
			pr = ti
		}
	}
	return res, nil
}

func (p *transcriber) skip(data *synthesizer.TTSDataPart) bool {
	return data.Cfg.JustAM || data.Cfg.Input.OutputFormat == api.AudioNone
}

func transWord(w *synthesizer.ProcessedWord) string {
	if w.TranscriptionWord != "" {
		return w.TranscriptionWord
	}
	return w.Tagged.Word
}

func mapTransOutput(data *synthesizer.TTSDataPart, out []transOutput) error {
	i := 0
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() {
			if len(out) <= i {
				return errors.New("wrong transcribe result")
			}
			err := setTrans(w, out[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

func setTrans(w *synthesizer.ProcessedWord, out transOutput) error {
	if out.Error != "" {
		return errors.Errorf("transcriber error for '%s'('%s'): %s", transWord(w), out.Word, out.Error)
	}
	if transWord(w) != out.Word {
		return errors.Errorf("words do not match (transcriber) '%s' vs '%s'", transWord(w), out.Word)
	}
	for _, t := range out.Transcription {
		if t.Transcription != "" {
			w.Transcription = dropQMarks(t.Transcription)
			return nil
		}
	}
	return nil
}

func dropQMarks(s string) string {
	return strings.ReplaceAll(s, "?", "")
}
