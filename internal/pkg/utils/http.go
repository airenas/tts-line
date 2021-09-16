package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//HTTPWrap for http call
type HTTPWrap struct {
	HTTPClient *http.Client
	URL        string
	Timeout    time.Duration
	flog       func(string, string)
}

//NewHTTWrap creates new wrapper
func NewHTTWrap(urlStr string) (*HTTPWrap, error) {
	return NewHTTWrapT(urlStr, time.Second*120)
}

//NewHTTWrapT creates new wrapper with timer
func NewHTTWrapT(urlStr string, timeout time.Duration) (*HTTPWrap, error) {
	res := &HTTPWrap{}
	var err error
	res.URL, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't parse url '%s'", urlStr)
	}
	res.HTTPClient = &http.Client{}
	res.Timeout = timeout
	res.flog = func(st, data string) { LogData(st, data) }
	return res, nil
}

//InvokeText makes http call with text
func (hw *HTTPWrap) InvokeText(dataIn string, dataOut interface{}) error {
	req, err := http.NewRequest("POST", hw.URL, strings.NewReader(dataIn))
	if err != nil {
		return errors.Wrapf(err, "Can't prepare request to '%s'", hw.URL)
	}
	hw.flog("Input : ", dataIn)
	req.Header.Set("Content-Type", "text/plain")
	return hw.invoke(req, dataOut)
}

//InvokeJSON makes http call with json
func (hw *HTTPWrap) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	return hw.InvokeJSONU(hw.URL, dataIn, dataOut)
}

//InvokeJSONU makes http call with json
func (hw *HTTPWrap) InvokeJSONU(URL string, dataIn interface{}, dataOut interface{}) error {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(dataIn)
	if err != nil {
		return err
	}
	hw.flog("Input : ", b.String())
	req, err := http.NewRequest("POST", URL, b)
	if err != nil {
		return errors.Wrapf(err, "Can't prepare request to '%s'", URL)
	}
	req.Header.Set("Content-Type", "application/json")
	return hw.invoke(req, dataOut)
}

func (hw *HTTPWrap) invoke(req *http.Request, dataOut interface{}) error {
	if hw.Timeout > 0 {
		ctx, cancelF := context.WithTimeout(context.Background(), hw.Timeout)
		defer cancelF()
		req = req.WithContext(ctx)
	}
	hw.flog("Call : ", req.URL.String())
	resp, err := hw.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Can't call '%s'", req.URL.String())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.Errorf("Can't invoke '%s'. Code: '%d'", req.URL.String(), resp.StatusCode)
	}
	br, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Can't read body")
	}
	hw.flog("Output: ", string(br))
	err = json.Unmarshal(br, dataOut)
	if err != nil {
		return errors.Wrap(err, "Can't decode response")
	}
	return nil
}
