package processor

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type transcriber struct {
	httpWrap HTTPInvokerJSON
}

//NewTranscriber creates new processor
func NewTranscriber(urlStr string) (synthesizer.PartProcessor, error) {
	res := &transcriber{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *transcriber) Process(data *synthesizer.TTSDataPart) error {
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
					return nil, errors.New("No accent variant for " + tword)
				}
				ti.Acc = w.AccentVariant.Accent
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
				return errors.New("Wrong transcribe result")
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
	if transWord(w) != out.Word {
		return errors.Errorf("Words do not match '%s' vs '%s'", transWord(w), out.Word)
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
