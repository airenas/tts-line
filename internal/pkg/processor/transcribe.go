package processor

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type transcriber struct {
	httpWrap HTTPInvokerJSON
}

// NewTranscriber creates new processor
func NewTranscriber(urlStr string) (synthesizer.PartProcessor, error) {
	res := &transcriber{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *transcriber) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "transcriber.Process")
	defer span.End()

	if skipTranscribe(data.Cfg) {
		log.Ctx(ctx).Info().Msg("Skip transcriber")
		return nil
	}

	inData, err := mapTransInput(data)
	if err != nil {
		return err
	}
	if len(inData) > 0 {
		var output []transOutput
		err := p.httpWrap.InvokeJSON(ctx, inData, &output)
		if err != nil {
			return err
		}
		err = mapTransOutput(data, output)
		if err != nil {
			return err
		}
	} else {
		log.Ctx(ctx).Debug().Msg("Skip transcriber - no data in")
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
			if !tgw.Space {
				pr = nil
			}
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
				ti.Syll = getSyllables(w)
				ti.User = getUserOEPal(w)
				if ti.User == "" {
					ti.Acc = synthesizer.GetTranscriberAccent(w)
					ti.Ml = w.AccentVariant.Ml
				}
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

func getSyllables(w *synthesizer.ProcessedWord) string {
	if w.TextPart != nil && w.TextPart.Accented != "" && w.TextPart.Syllables != "" { // provided by user
		return w.TextPart.Syllables
	}
	if w.AccentVariant != nil {
		return w.AccentVariant.Syll
	}
	return ""
}

func getUserOEPal(w *synthesizer.ProcessedWord) string {
	res := ""
	if w.TextPart != nil && w.TextPart.Accented != "" && w.TextPart.UserOEPal != "" { // provided by user
		res = w.TextPart.UserOEPal
		if w.UserAccent > 0 {
			res = addAccent(res, w.UserAccent)
		}
	}
	return res
}

func addAccent(in string, acc int) string {
	res := strings.Builder{}
	acc, pos, li := acc/100, acc%100, 0
	for _, s := range in {
		res.WriteRune(s)
		if unicode.IsLetter(s) {
			li++
			if li == pos {
				res.WriteString(accent.TranscriberAccent(acc))
			}
		}
	}
	return res.String()
}

func skipTranscribe(cfg *synthesizer.TTSConfig) bool {
	return cfg.JustAM || (cfg.Input.OutputFormat == api.AudioNone && cfg.Input.OutputTextFormat != api.TextTranscribed)
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
