package processor

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type cliticDetector struct {
	httpWrap HTTPInvokerJSON
}

// NewClitics creates new processor
func NewClitics(urlStr string) (synthesizer.PartProcessor, error) {
	res := &cliticDetector{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *cliticDetector) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip clitics")
		return nil
	}
	inData, err := mapCliticsInput(data)
	if err != nil {
		return err
	}
	if len(inData) > 0 {
		var output []api.CliticsOutput
		err := p.httpWrap.InvokeJSON(ctx, inData, &output)
		if err != nil {
			return err
		}
		err = mapCliticsOutput(data, output)
		if err != nil {
			return err
		}
	} else {
		log.Ctx(ctx).Debug().Msg("Skip clitics - no data in")
	}
	return nil
}

func mapCliticsInput(data *synthesizer.TTSDataPart) ([]*api.CliticsInput, error) {
	res := []*api.CliticsInput{}
	for i, w := range data.Words {
		tgw := w.Tagged
		ci := &api.CliticsInput{}
		ci.ID = i
		ci.String = strings.ToLower(transWord(w))
		ci.Lemma = tgw.Lemma
		ci.Mi = tgw.Mi
		ci.Type = toType(&w.Tagged)
		res = append(res, ci)
	}
	return res, nil
}

func toType(t *synthesizer.TaggedWord) string {
	if t.IsWord() {
		return "WORD"
	}
	if t.Space {
		return "SPACE"
	}
	return "OTHER"
}

func mapCliticsOutput(data *synthesizer.TTSDataPart, out []api.CliticsOutput) error {
	for _, co := range out {
		if co.ID >= len(data.Words) {
			return errors.Errorf("wrong clitics output ID = '%d'. Max %d", co.ID, len(data.Words))
		}
		w := data.Words[co.ID]
		if co.AccentType == api.TypeStatic {
			w.Clitic.Type = synthesizer.CliticsCustom
			w.Clitic.Accent = co.Accent
		} else if co.AccentType == api.TypeNone {
			w.Clitic.Type = synthesizer.CliticsNone
		} else {
			w.Clitic.Type = synthesizer.CliticsUnused
		}
	}
	return nil
}

func (p *cliticDetector) skip(data *synthesizer.TTSDataPart) bool {
	return data.Cfg.JustAM
}
