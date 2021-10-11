package wrapservice

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
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
	goapp.Log.Infof("AM URL: %s", amURL+"/model")
	am, err := utils.NewHTTPWrapT(amURL+"/model", time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
	}
	amService, err := utils.NewHTTPBackoff(am, newBackoff)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
	}
	amService.InvokeIndicatorFunc = func(d interface{}) {
		totalInvokeMetrics.WithLabelValues("am", d.(*amInput).Voice).Add(1)
	}
	amService.RetryIndicatorFunc = func(d interface{}) {
		totalRetryMetrics.WithLabelValues("am", d.(*amInput).Voice).Add(1)
	}
	res.amWrap = amService

	goapp.Log.Infof("Vocoder URL: %s", vocURL+"/model")
	voc, err := utils.NewHTTPWrapT(vocURL+"/model", time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, "can't init Vocoder client")
	}
	vocService, err := utils.NewHTTPBackoff(voc, newBackoff)
	if err != nil {
		return nil, errors.Wrap(err, "can't init Vocoder client")
	}
	vocService.InvokeIndicatorFunc = func(d interface{}) {
		totalInvokeMetrics.WithLabelValues("vocoder", d.(*vocInput).Voice).Add(1)
	}
	vocService.RetryIndicatorFunc = func(d interface{}) {
		totalRetryMetrics.WithLabelValues("vocoder", d.(*vocInput).Voice).Add(1)
	}
	res.vocWrap = vocService
	return res, nil
}

//Work is main method
func (p *Processor) Work(text string, speed float32, voice string) (string, error) {
	amIn := amInput{Text: text, Speed: speed, Voice: voice}
	var amOut output
	err := p.amWrap.InvokeJSON(&amIn, &amOut)
	if err != nil {
		totalFailureMetrics.WithLabelValues("am", voice).Add(1)
		return "", errors.Wrap(err, "can't invoke AM")
	}
	vocIn := vocInput{Data: amOut.Data, Voice: voice}
	var vocOut output
	err = p.vocWrap.InvokeJSON(&vocIn, &vocOut)
	if err != nil {
		totalFailureMetrics.WithLabelValues("vocoder", voice).Add(1)
		return "", errors.Wrap(err, "can't invoke Vocoder")
	}
	return vocOut.Data, nil
}

type amInput struct {
	Text  string  `json:"text"`
	Speed float32 `json:"speedAlpha,omitempty"`
	Voice string  `json:"voice,omitempty"`
}

type vocInput struct {
	Data  string `json:"data"`
	Voice string `json:"voice,omitempty"`
}

type output struct {
	Data string `json:"data"`
}

func newBackoff() backoff.BackOff {
	res := backoff.NewExponentialBackOff()
	res.InitialInterval = time.Second * 2
	return backoff.WithMaxRetries(res, 3)
}
