package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/pkg/errors"
)

//HTTPWrap for http call
type HTTPWrap struct {
	HTTPClient *http.Client
	URL        string
}

//NewHTTWrap creates new wrapper
func NewHTTWrap(urlStr string) (*HTTPWrap, error) {
	res := &HTTPWrap{}
	var err error
	res.URL, err = checkURL(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url")
	}
	res.HTTPClient = &http.Client{}
	return res, nil
}

//InvokeText makes http call with text
func (hw *HTTPWrap) InvokeText(dataIn string, dataOut interface{}) error {
	req, err := http.NewRequest("POST", hw.URL, strings.NewReader(dataIn))
	if err != nil {
		return err
	}
	logString("Input : ", dataIn)
	req.Header.Set("Content-Type", "text/plain")
	return hw.invoke(req, dataOut)
}

//InvokeJSON makes http call with json
func (hw *HTTPWrap) InvokeJSON(dataIn interface{}, dataOut interface{}) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(dataIn)
	if err != nil {
		return err
	}
	logString("Input : ", b.String())
	req, err := http.NewRequest("POST", hw.URL, b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return hw.invoke(req, dataOut)
}

func (hw *HTTPWrap) invoke(req *http.Request, dataOut interface{}) error {
	logString("Call : ", hw.URL)
	resp, err := hw.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("Can't invoke")
	}
	br, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Can't read body")
	}
	logString("Output: ", string(br))
	err = json.Unmarshal(br, dataOut)
	if err != nil {
		return errors.Wrap(err, "Can't decode response")
	}
	return nil
}

//MaxLogDataSize indicates how many bytes of data to log
var MaxLogDataSize = 100

func logString(st string, data string) {
	if len(data) > MaxLogDataSize {
		data = data[0:MaxLogDataSize] + "..."
	}
	goapp.Log.Debugf("%s %s", st, data)
}
