package processor

import (
	"context"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type audioConverter struct {
	httpWrap HTTPInvokerJSON
}

// NewConverter creates new processor for wav to mp3/m4a conversion
func NewConverter(urlStr string) (synthesizer.Processor, error) {
	res := &audioConverter{}
	var err error
	res.httpWrap, err = utils.NewHTTPWrapT(urlStr, time.Minute*5)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *audioConverter) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "audioConverter.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	inData := audioConvertInput{Data: data.Audio, Format: data.Input.OutputFormat.String(),
		Metadata: data.Input.OutputMetadata}
	var output audioConvertOutput
	err := p.httpWrap.InvokeJSON(ctx, inData, &output)
	if err != nil {
		return err
	}
	data.AudioMP3 = output.Data
	return nil
}

// Info return info about processor
func (p *audioConverter) Info() string {
	return "audioConverter"
}

type audioConvertInput struct {
	Data     string   `json:"audio"`
	Format   string   `json:"format"`
	Metadata []string `json:"metadata"`
}

type audioConvertOutput struct {
	Data string `json:"audio"`
}
