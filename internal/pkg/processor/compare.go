package processor

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type comparator struct {
	httpWrap HTTPInvokerJSON
}

//NewComparator creates new text comparator processor
func NewComparator(urlStr string) (synthesizer.Processor, error) {
	res := &comparator{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *comparator) Process(data *synthesizer.TTSData) error {
	defer goapp.Estimate("Compare")()
	utils.LogData("Input: ", data.OriginalText)
	utils.LogData("Input previous: ", data.PreviousText)
	// inData := &normData{Text: data.OriginalText}
	// var output normData
	// err := p.httpWrap.InvokeJSON(inData, &output)
	// if err != nil {
	// 	return err
	// }

	return nil
}
