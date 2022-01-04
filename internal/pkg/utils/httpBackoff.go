package utils

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

type HTTPInvokerJSON interface {
	InvokeJSONU(string, interface{}, interface{}) error
	InvokeJSON(interface{}, interface{}) error
}

//HTTPBackof http call with backoff
type HTTPBackoff struct {
	HTTPClient          HTTPInvokerJSON
	backoffF            func() backoff.BackOff
	InvokeIndicatorFunc func(interface{})
	RetryIndicatorFunc  func(interface{})
}

//NewHTTPBackoff creates new wrapper with backoff
func NewHTTPBackoff(realWrapper HTTPInvokerJSON, backoffF func() backoff.BackOff) (*HTTPBackoff, error) {
	return &HTTPBackoff{HTTPClient: realWrapper, backoffF: backoffF}, nil
}

//InvokeJSON makes http call with json
func (hw *HTTPBackoff) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	return hw.invoke(func() error {
		return hw.HTTPClient.InvokeJSON(dataIn, dataOut)
	}, dataIn)
}

func (hw *HTTPBackoff) InvokeJSONU(URL string, dataIn interface{}, dataOut interface{}) error {
	return hw.invoke(func() error {
		return hw.HTTPClient.InvokeJSONU(URL, dataIn, dataOut)
	}, dataIn)
}

func (hw *HTTPBackoff) invoke(f func() error, dataIn interface{}) error {
	failC := 0
	op := func() error {
		if hw.InvokeIndicatorFunc != nil {
			hw.InvokeIndicatorFunc(dataIn)
		}
		if failC > 0 && hw.RetryIndicatorFunc != nil {
			hw.RetryIndicatorFunc(dataIn)
		}
		err := f()
		if err != nil {
			failC++
			goapp.Log.Error(errors.Wrapf(err, "failed %d time(s)", failC))
			return err
		}
		return nil
	}
	err := backoff.Retry(op, hw.backoffF())
	if err == nil && failC > 0 {
		goapp.Log.Infof("Success after retrying %d time(s)", failC)
	}
	return err
}
