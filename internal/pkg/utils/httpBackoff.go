package utils

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/cenkalti/backoff/v4"
)

type HTTPInvokerJSON interface {
	InvokeJSON(interface{}, interface{}) error
}

//HTTPBackof http call with backoff
type HTTPBackoff struct {
	HTTPClient HTTPInvokerJSON
	backoffF   func() backoff.BackOff
}

//NewHTTPBackoff creates new wrapper with backoff
func NewHTTPBackoff(realWrapper HTTPInvokerJSON, backoffF func() backoff.BackOff) (*HTTPBackoff, error) {
	return &HTTPBackoff{HTTPClient: realWrapper, backoffF: backoffF}, nil
}

//InvokeJSON makes http call with json
func (hw *HTTPBackoff) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	failC := 0
	op := func() error {
		err := hw.HTTPClient.InvokeJSON(dataIn, dataOut)
		if err != nil {
			goapp.Log.Error(err)
			failC++
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
