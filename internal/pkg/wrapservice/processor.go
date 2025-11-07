package wrapservice

import (
	"context"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/processor"
	"github.com/airenas/tts-line/internal/pkg/syntmodel"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

// Processor does synthesis work
type Processor struct {
	amWrap  processor.HTTPInvokerJSON
	vocWrap processor.HTTPInvokerJSON
}

// NewProcessor creates new processor
func NewProcessor(amURL, vocURL string) (*Processor, error) {
	res := &Processor{}
	goapp.Log.Info().Msgf("AM URL: %s", amURL+"/model")
	am, err := utils.NewHTTPWrapT(amURL+"/model", time.Minute*2)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
	}
	am = am.WithOutputFormat(utils.EncodingFormatMsgPack)
	amService, err := utils.NewHTTPBackoff(am, newBackoff, utils.RetryAll)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
	}
	amService.InvokeIndicatorFunc = func(d interface{}) {
		totalInvokeMetrics.WithLabelValues("am", d.(*syntmodel.AMInput).Voice).Add(1)
	}
	amService.RetryIndicatorFunc = func(d interface{}) {
		totalRetryMetrics.WithLabelValues("am", d.(*syntmodel.AMInput).Voice).Add(1)
	}
	res.amWrap = amService

	goapp.Log.Info().Msgf("Vocoder URL: %s", vocURL+"/model")
	voc, err := utils.NewHTTPWrapT(vocURL+"/model", time.Minute*2)
	if err != nil {
		return nil, errors.Wrap(err, "can't init Vocoder client")
	}
	voc = voc.WithInputFormat(utils.EncodingFormatMsgPack).WithOutputFormat(utils.EncodingFormatMsgPack)
	vocService, err := utils.NewHTTPBackoff(voc, newBackoff, utils.RetryAll)
	if err != nil {
		return nil, errors.Wrap(err, "can't init Vocoder client")
	}
	vocService.InvokeIndicatorFunc = func(d interface{}) {
		totalInvokeMetrics.WithLabelValues("vocoder", d.(*syntmodel.VocInput).Voice).Add(1)
	}
	vocService.RetryIndicatorFunc = func(d interface{}) {
		totalRetryMetrics.WithLabelValues("vocoder", d.(*syntmodel.VocInput).Voice).Add(1)
	}
	res.vocWrap = vocService
	return res, nil
}

// Work is main method
func (p *Processor) Work(ctx context.Context, params *api.Params) (*syntmodel.Result, error) {
	ctx, span := utils.StartSpan(ctx, "Processor.Work")
	defer span.End()

	amIn := syntmodel.AMInput{Text: params.Text, Speed: params.Speed, Voice: params.Voice, Priority: params.Priority,
		DurationsChange: params.DurationsChange, PitchChange: params.PitchChange}
	var amOut syntmodel.AMOutput
	err := p.amWrap.InvokeJSON(ctx, &amIn, &amOut)
	if err != nil {
		totalFailureMetrics.WithLabelValues("am", params.Voice).Add(1)
		return nil, errors.Wrap(err, "can't invoke AM")
	}
	vocIn := syntmodel.VocInput{Data: amOut.Data, Voice: params.Voice, Priority: params.Priority}
	var vocOut syntmodel.VocOutput
	err = p.vocWrap.InvokeJSON(ctx, &vocIn, &vocOut)
	if err != nil {
		totalFailureMetrics.WithLabelValues("vocoder", params.Voice).Add(1)
		return nil, errors.Wrap(err, "can't invoke Vocoder")
	}
	return &syntmodel.Result{Data: vocOut.Data, Durations: amOut.Durations, SilDuration: amOut.SilDuration, Step: amOut.Step}, nil
}

func newBackoff() backoff.BackOff {
	res := backoff.NewExponentialBackOff()
	res.InitialInterval = time.Second * 2
	return backoff.WithMaxRetries(res, 3)
}
