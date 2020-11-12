package processor

import (
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type mp3Converter struct {
	httpWrap HTTPInvokerJSON
}

//NewMP3 creates new processor
func NewMP3(urlStr string) (synthesizer.Processor, error) {
	res := &mp3Converter{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *mp3Converter) Process(data *synthesizer.TTSData) error {
	inData := mp3Input{Data: data.Audio}
	var output mp3Output
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}
	data.AudioMP3 = output.Data
	return nil
}

type mp3Input struct {
	Data string `json:"audio"`
}

type mp3Output struct {
	Data string `json:"audio"`
}
