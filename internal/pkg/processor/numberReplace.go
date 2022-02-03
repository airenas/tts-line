package processor

import (
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

//HTTPInvoker makes http call
type HTTPInvoker interface {
	InvokeText(string, interface{}) error
}

type numberReplace struct {
	httpWrap HTTPInvoker
}

//NewNumberReplace creates new processor
func NewNumberReplace(urlStr string) (synthesizer.Processor, error) {
	res := &numberReplace{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *numberReplace) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip numberReplace")
		return nil
	}
	return p.httpWrap.InvokeText(data.Text, &data.TextWithNumbers)
}

func (p *numberReplace) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
