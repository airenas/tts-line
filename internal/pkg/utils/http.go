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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPWrap for http call
type HTTPWrap struct {
	HTTPClient *http.Client
	URL        string
	Timeout    time.Duration
	flog       func(context.Context, string, string, error)
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
	res.flog = func(ctx context.Context, st, data string, err error) { LogData(ctx, st, data, err) }
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
func (hw *HTTPWrap) InvokeText(ctx context.Context, dataIn string, dataOut interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hw.URL, strings.NewReader(dataIn))
	if err != nil {
		return errors.Wrapf(err, "can't prepare request to '%s'", hw.URL)
	}
	hw.flog(ctx, "Input", dataIn, nil)
	req.Header.Set("Content-Type", "text/plain")
	err = hw.invoke(ctx, req, dataOut)
	if err != nil {
		hw.flog(ctx, "Input", dataIn, err)
	}
	return err
}

// InvokeJSON makes http call with json
func (hw *HTTPWrap) InvokeJSON(ctx context.Context, dataIn interface{}, dataOut interface{}) error {
	return hw.InvokeJSONU(ctx, hw.URL, dataIn, dataOut)
}

// InvokeJSONU makes http call to URL with JSON
func (hw *HTTPWrap) InvokeJSONU(ctx context.Context, URL string, dataIn interface{}, dataOut interface{}) error {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(dataIn)
	if err != nil {
		return err
	}
	hw.flog(ctx, "Input", b.String(), nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, b)
	if err != nil {
		return errors.Wrapf(err, "can't prepare request to '%s'", URL)
	}
	req.Header.Set("Content-Type", "application/json")
	err = hw.invoke(ctx, req, dataOut)
	if err != nil {
		hw.flog(ctx, "Input", b.String(), err)
	}
	return err
}

func (hw *HTTPWrap) invoke(ctx context.Context, req *http.Request, dataOut interface{}) error {
	ctx, span := StartSpan(ctx, "HTTPWrap.invoke", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if hw.Timeout > 0 {
		ctx, cancelF := context.WithTimeout(ctx, hw.Timeout)
		defer cancelF()
		req = req.WithContext(ctx)
	}
	hw.flog(ctx, "Call", req.URL.String(), nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
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
	hw.flog(ctx, "Output", string(br), nil)
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
