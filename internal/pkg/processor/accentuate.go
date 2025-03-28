package processor

import (
	"context"
	"time"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type accentuator struct {
	httpWrap HTTPInvokerJSON
}

// NewAccentuator creates new processor
func NewAccentuator(urlStr string) (synthesizer.PartProcessor, error) {
	res := &accentuator{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *accentuator) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "accentuator.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip accentuator")
		return nil
	}
	inData := mapAccentInput(data)
	if len(inData) > 0 {

		var output []accentOutputElement
		err := p.httpWrap.InvokeJSON(ctx, inData, &output)
		if err != nil {
			return err
		}
		err = mapAccentOutput(data, output)
		if err != nil {
			return err
		}
	} else {
		log.Ctx(ctx).Debug().Msg("Skip accenter - no data in")
	}
	return nil
}

type accentOutputElement struct {
	Accent []accentInfo `json:"accent"`
	Word   string       `json:"word"`
	Error  string       `json:"error"`
}

type accentInfo struct {
	MF       string                      `json:"mf"`
	Mi       string                      `json:"mi"`
	MiVdu    string                      `json:"mi_vdu"`
	Mih      string                      `json:"mih"`
	Error    string                      `json:"error"`
	Variants []synthesizer.AccentVariant `json:"variants"`
}

func mapAccentInput(data *synthesizer.TTSDataPart) []string {
	res := []string{}
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && w.UserTranscription == "" {
			res = append(res, w.Tagged.Word)
		}
	}
	return res
}

func mapAccentOutput(data *synthesizer.TTSDataPart, out []accentOutputElement) error {
	i := 0
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && w.UserTranscription == "" {
			if len(out) <= i {
				return errors.New("wrong accent result")
			}
			err := setAccent(w, out[i])
			if err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

func setAccent(w *synthesizer.ProcessedWord, out accentOutputElement) error {
	if out.Error != "" {
		if len(w.Tagged.Word) >= 50 {
			return utils.NewErrWordTooLong(w.Tagged.Word)
		}
		if len(w.Tagged.Word) > 0 && out.Error == "No word" {
			return utils.NewErrBadSymbols(w.Tagged.Word, "")
		}
		return errors.Errorf("accent error for '%s'('%s'): %s", w.Tagged.Word, out.Word, out.Error)
	}
	if w.Tagged.Word != out.Word {
		return utils.NewErrBadSymbols(w.Tagged.Word, out.Word)
	}
	w.AccentVariant = findBestAccentVariant(out.Accent, w.Tagged.Mi, w.Tagged.Lemma)
	return nil
}

func findBestAccentVariant(acc []accentInfo, mi string, lema string) *synthesizer.AccentVariant {
	find := func(fa func(a *accentInfo) bool, fv func(v *synthesizer.AccentVariant) bool) *synthesizer.AccentVariant {
		for _, a := range acc {
			if fa(&a) {
				for _, v := range a.Variants {
					if fv(&v) {
						return &v
					}
				}
			}
		}
		return nil
	}
	fIsAccent := func(v *synthesizer.AccentVariant) bool { return v.Accent > 0 }

	if res := find(func(a *accentInfo) bool { return a.Error == "" && a.MiVdu == mi && a.MF == lema }, fIsAccent); res != nil {
		return res
	}

	if res := find(func(a *accentInfo) bool { return a.Error == "" && a.MiVdu == mi }, fIsAccent); res != nil {
		return res
	}
	// no mi filter
	if res := find(func(a *accentInfo) bool { return a.Error == "" }, fIsAccent); res != nil {
		return res
	}
	//no filter
	return find(func(a *accentInfo) bool { return true }, func(v *synthesizer.AccentVariant) bool { return true })
}

func (p *accentuator) skip(data *synthesizer.TTSDataPart) bool {
	return data.Cfg.JustAM
}

func newHTTPWrapBackoff(urlStr string, timeout time.Duration) (HTTPInvokerJSON, error) {
	real, err := utils.NewHTTPWrapT(urlStr, timeout)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http wrap")
	}
	res, err := utils.NewHTTPBackoff(real, newSimpleBackoff, utils.IsRetryable)
	if err != nil {
		return nil, errors.Wrap(err, "can't init backoff http client")
	}
	return res, nil
}

func newSimpleBackoff() backoff.BackOff {
	res := backoff.NewExponentialBackOff()
	return backoff.WithMaxRetries(res, 4)
}
