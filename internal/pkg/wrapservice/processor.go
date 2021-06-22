package wrapservice

import (
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

//Processor does synthesis work
type Processor struct {
	amWrap  processor.HTTPInvokerJSON
	vocWrap processor.HTTPInvokerJSON
}

//NewProcessor creates new processor
func NewProcessor(amURL, vocURL string) (*Processor, error) {
	res := &Processor{}
	var err error
	res.amWrap, err = utils.NewHTTWrap(amURL)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init AM client")
	}
	res.vocWrap, err = utils.NewHTTWrap(vocURL)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init Vocoder client")
	}
	return res, nil
}

//Work is main method
func (p *Processor) Work(text string, speed float32) (string, error) {
	amIn := amInput{Text: text, Speed: speed}
	var amOut output
	err := p.amWrap.InvokeJSON(&amIn, &amOut)
	if err != nil {
		return "", errors.Wrap(err, "Can't invoke AM")
	}
	vocIn := output{Data: amOut.Data}
	var vocOut output
	err = p.vocWrap.InvokeJSON(&vocIn, &vocOut)
	if err != nil {
		return "", errors.Wrap(err, "Can't invoke Vocoder")
	}
	return vocOut.Data, nil
}

type amInput struct {
	Text  string  `json:"text"`
	Speed float32 `json:"speedAlpha,omitempty"`
}

type output struct {
	Data string `json:"data"`
}
