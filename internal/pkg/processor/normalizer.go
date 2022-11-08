package processor

import (
	"fmt"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type normalizer struct {
	httpWrap HTTPInvokerJSON
}

//NewNormalizer creates new text normalize processor
func NewNormalizer(urlStr string) (synthesizer.Processor, error) {
	res := &normalizer{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *normalizer) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip normalize")
		return nil
	}
	defer goapp.Estimate("Normalize")()
	txt := data.CleanedText
	utils.LogData("Input: ", txt)
	inData := &normRequestData{Orig: txt}
	var output normResponseData
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}

	data.NormalizedText = output.Res
	utils.LogData("Output: ", data.NormalizedText)
	return nil
}

type normRequestData struct {
	Orig string `json:"org"`
}

type normResponseData struct {
	Err string                `json:"err"`
	Org string                `json:"org"`
	Rep []normResponseDataRep `json:"rep"`
	Res string                `json:"res"`
}

type normResponseDataRep struct {
	Beg  int64  `json:"beg"`
	End  int64  `json:"end"`
	Text string `json:"text"`
}

func (p *normalizer) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *normalizer) Info() string {
	return fmt.Sprintf("normalizer(%s)", utils.RetrieveInfo(p.httpWrap))
}
