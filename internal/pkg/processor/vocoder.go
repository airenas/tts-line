package processor

import (
	"context"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type vocoder struct {
	httpWrap HTTPInvokerJSON
	url      string
}

// NewVocoder creates new processor
func NewVocoder(urlStr string) (synthesizer.PartProcessor, error) {
	res := &vocoder{}

	res.url = urlStr
	voc, err := utils.NewHTTPWrapT(getVoiceURL(res.url, "testVoice"), time.Second*120)
	if err != nil {
		return nil, errors.Wrap(err, "can't init vocoder client")
	}
	res.httpWrap, err = utils.NewHTTPBackoff(voc, newGPUBackoff, utils.RetryAll)
	if err != nil {
		return nil, errors.Wrap(err, "can't init vocoder client")
	}

	return res, nil
}

func (p *vocoder) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "vocoder.Process")
	defer span.End()

	if data.Cfg.Input.OutputFormat == api.AudioNone {
		return nil
	}

	inData := vocInput{Data: data.Spectogram, Voice: data.Cfg.Input.Voice, Priority: data.Cfg.Input.Priority}
	var output vocOutput
	err := p.httpWrap.InvokeJSONU(ctx, getVoiceURL(p.url, data.Cfg.Input.Voice), inData, &output)
	if err != nil {
		return err
	}
	data.Audio = output.Data
	return nil
}

type vocInput struct {
	Data     string `json:"data"`
	Voice    string `json:"voice"`
	Priority int    `json:"priority,omitempty"`
}

type vocOutput struct {
	Data string `json:"data"`
}
