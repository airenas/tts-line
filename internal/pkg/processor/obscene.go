package processor

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type obscene struct {
	httpWrap HTTPInvokerJSON
}

// NewObsceneFilter creates new processor
func NewObsceneFilter(urlStr string) (synthesizer.PartProcessor, error) {
	res := &obscene{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *obscene) Process(data *synthesizer.TTSDataPart) error {
	if p.skip(data) {
		goapp.Log.Info().Msg("Skip obscene filter")
		return nil
	}
	inData := mapObsceneInput(data)
	if len(inData) > 0 {

		var output []obsceneResultToken
		err := p.httpWrap.InvokeJSON(inData, &output)
		if err != nil {
			return err
		}
		err = mapObsceneOutput(data, output)
		if err != nil {
			return err
		}
	} else {
		goapp.Log.Debug().Msg("Skip obscene filter - no data in")
	}
	return nil
}

type obsceneToken struct {
	Token string `json:"token"`
}

type obsceneResultToken struct {
	Token   string `json:"token"`
	Obscene int    `json:"obscene"`
}

func mapObsceneInput(data *synthesizer.TTSDataPart) []*obsceneToken {
	res := []*obsceneToken{}
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && w.UserTranscription == "" {
			res = append(res, &obsceneToken{Token: w.Tagged.Word})
		}
	}
	return res
}

func mapObsceneOutput(data *synthesizer.TTSDataPart, out []obsceneResultToken) error {
	i := 0
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.IsWord() && w.UserTranscription == "" {
			if len(out) <= i {
				return errors.Errorf("wrong obscene filter result. Index %d, len(out) = %d", i, len(out))
			}
			if w.Tagged.Word != out[i].Token {
				return errors.Errorf("wrong obscene filter result. Index %d, wanted %s, got %s",
					i, w.Tagged.Word, out[i].Token)
			}
			w.Obscene = out[i].Obscene == 1
			i++
		}
	}
	return nil
}

func (p *obscene) skip(data *synthesizer.TTSDataPart) bool {
	return data.Cfg.JustAM
}
