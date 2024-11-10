package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

// HTTPWrap for http call
type HTTPWrap struct {
	HTTPClient *http.Client
	URL        string
	Timeout    time.Duration
	flog       func(string, string, error)
}

// NewHTTPWrap creates new wrapper
func NewHTTPWrap(urlStr string) (*HTTPWrap, error) {
	return NewHTTPWrapT(urlStr, time.Second*120)
}

// NewHTTPWrapT creates new wrapper with timer
func NewHTTPWrapT(urlStr string, timeout time.Duration) (*HTTPWrap, error) {
	res := &HTTPWrap{}
	var err error
	res.URL, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse url '%s'", urlStr)
	}
	res.HTTPClient = &http.Client{Transport: newTransport()}
	res.Timeout = timeout
	res.flog = func(st, data string, err error) { LogData(st, data, err) }
	return res, nil
}

func newTransport() http.RoundTripper {
	res := http.DefaultTransport.(*http.Transport).Clone()
	res.MaxIdleConns = 40
	res.MaxConnsPerHost = 100
	res.IdleConnTimeout = 90 * time.Second
	res.MaxIdleConnsPerHost = 20
	return res
}

// InvokeText makes http call with text
func (hw *HTTPWrap) InvokeText(dataIn string, dataOut interface{}) error {
	req, err := http.NewRequest(http.MethodPost, hw.URL, strings.NewReader(dataIn))
	if err != nil {
		return errors.Wrapf(err, "can't prepare request to '%s'", hw.URL)
	}
	hw.flog("Input", dataIn, nil)
	req.Header.Set("Content-Type", "text/plain")
	err = hw.invoke(req, dataOut)
	if err != nil {
		hw.flog("Input", dataIn, err)
	}
	return err
}

// InvokeJSON makes http call with json
func (hw *HTTPWrap) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	return hw.InvokeJSONU(hw.URL, dataIn, dataOut)
}

// InvokeJSONU makes http call to URL with JSON
func (hw *HTTPWrap) InvokeJSONU(URL string, dataIn interface{}, dataOut interface{}) error {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(dataIn)
	if err != nil {
		return err
	}
	hw.flog("Input", b.String(), nil)
	req, err := http.NewRequest(http.MethodPost, URL, b)
	if err != nil {
		return errors.Wrapf(err, "can't prepare request to '%s'", URL)
	}
	req.Header.Set("Content-Type", "application/json")
	err = hw.invoke(req, dataOut)
	if err != nil {
		hw.flog("Input", b.String(), err)
	}
	return err
}

func (hw *HTTPWrap) invoke(req *http.Request, dataOut interface{}) error {
	if hw.Timeout > 0 {
		ctx, cancelF := context.WithTimeout(context.Background(), hw.Timeout)
		defer cancelF()
		req = req.WithContext(ctx)
	}
	hw.flog("Call", req.URL.String(), nil)
	resp, err := hw.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "can't call '%s'", req.URL.String())
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 10000))
		_ = resp.Body.Close()
	}()

	if err := goapp.ValidateHTTPResp(resp, 100); err != nil {
		return errors.Wrapf(err, "can't invoke '%s'", req.URL.String())
	}
	br, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "can't read body")
	}
	hw.flog("Output", string(br), nil)
	err = json.Unmarshal(br, dataOut)
	if err != nil {
		return errors.Wrap(err, "can't decode response")
	}
	return nil
}

// Info returns info about wrapper
func (hw *HTTPWrap) Info() string {
	return fmt.Sprintf("HTTPWrap(%s, tm: %s)", hw.URL, hw.Timeout.String())
}
