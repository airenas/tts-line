package utils

import (
	"context"
	"io"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

//HTTPInvokerJSON http POST invoker with JSON in, out params
type HTTPInvokerJSON interface {
	InvokeJSONU(string, interface{}, interface{}) error
	InvokeJSON(interface{}, interface{}) error
	InvokeText(string, interface{}) error
}

//HTTPBackoff http call with backoff
type HTTPBackoff struct {
	HTTPClient          HTTPInvokerJSON
	backoffF            func() backoff.BackOff
	errF                func(error) error
	InvokeIndicatorFunc func(interface{})
	RetryIndicatorFunc  func(interface{})
}

//NewHTTPBackoff creates new wrapper with backoff
//errF must return same error or wrap to backoff.PermanentError to stop backoff
func NewHTTPBackoff(realWrapper HTTPInvokerJSON, backoffF func() backoff.BackOff, errF func(error) error) (*HTTPBackoff, error) {
	return &HTTPBackoff{HTTPClient: realWrapper, backoffF: backoffF, errF: errF}, nil
}

//InvokeJSON makes http call with json
func (hw *HTTPBackoff) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	return hw.invoke(func() error {
		return hw.HTTPClient.InvokeJSON(dataIn, dataOut)
	}, dataIn)
}

//InvokeText makes http call with text input
func (hw *HTTPBackoff) InvokeText(dataIn string, dataOut interface{}) error {
	return hw.invoke(func() error {
		return hw.HTTPClient.InvokeText(dataIn, dataOut)
	}, dataIn)
}

//InvokeJSONU makes call to URL wits JSON
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
			goapp.Log.Warn(errors.Wrapf(err, "failed %d time(s)", failC))
			return hw.errF(err)
		}
		return nil
	}
	err := backoff.Retry(op, hw.backoffF())
	if err == nil && failC > 0 {
		goapp.Log.Infof("Success after retrying %d time(s)", failC)
	}
	return err
}

//RetryAll - error map function for NewHTTPBackoff to retry all errors
func RetryAll(err error) error {
	return err
}

//RetryEOF - error map function for NewHTTPBackoff to retry io.EOF and timeout errors
func RetryEOF(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return &backoff.PermanentError{Err: err}
}
