package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

//HTTPWrap for http call
type HTTPWrap struct {
	HTTPClient *http.Client
	URL        string
	flog       func(string, string)
}

//NewHTTWrap creates new wrapper
func NewHTTWrap(urlStr string) (*HTTPWrap, error) {
	res := &HTTPWrap{}
	var err error
	res.URL, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't parse url '%s'", urlStr)
	}
	res.HTTPClient = &http.Client{}
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
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	err := enc.Encode(dataIn)
	if err != nil {
		return err
	}
	hw.flog("Input : ", b.String())
	req, err := http.NewRequest("POST", hw.URL, b)
	if err != nil {
		return errors.Wrapf(err, "Can't prepare request to '%s'", hw.URL)
	}
	req.Header.Set("Content-Type", "application/json")
	return hw.invoke(req, dataOut)
}

func (hw *HTTPWrap) invoke(req *http.Request, dataOut interface{}) error {
	hw.flog("Call : ", hw.URL)
	resp, err := hw.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Can't call '%s'", hw.URL)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.Errorf("Can't invoke '%s'. Code: '%d'", hw.URL, resp.StatusCode)
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
