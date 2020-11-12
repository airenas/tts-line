package processor

import (
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type vocoder struct {
	httpWrap HTTPInvokerJSON
}

//NewVocoder creates new processor
func NewVocoder(urlStr string) (synthesizer.Processor, error) {
	res := &vocoder{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *vocoder) Process(data *synthesizer.TTSData) error {
	inData := vocInput{Data: data.Spectogram}
	var output *vocOutput
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}
	data.Audio = output.Data
	return nil
}

type vocInput struct {
	Data string `json:"data"`
}

type vocOutput struct {
	Data string `json:"data"`
}
