package wrapservice

import (
	"time"

	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/cenkalti/backoff/v4"
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
	am, err := utils.NewHTTWrapT(amURL, time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init AM client")
	}
	res.amWrap, err = utils.NewHTTPBackoff(am, newBackoff)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init AM client")
	}
	voc, err := utils.NewHTTWrapT(vocURL, time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init Vocoder client")
	}
	res.vocWrap, err = utils.NewHTTPBackoff(voc, newBackoff)
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

func newBackoff() backoff.BackOff {
	res := backoff.NewExponentialBackOff()
	res.InitialInterval = time.Second * 2
	return backoff.WithMaxRetries(res, 3)
}
