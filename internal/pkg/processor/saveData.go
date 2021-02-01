package processor

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

//SaverDB interface for text saving
type SaverDB interface {
	Save(req, text string, reqType utils.RequestTypeEnum) error
}

type saver struct {
	s SaverDB
}

//NewSaver creates new text to db processor
func NewSaver(s SaverDB) (synthesizer.Processor, error) {
	if s == nil {
		return nil, errors.New("No Saver")
	}
	return &saver{s: s}, nil
}

func (p *saver) Process(data *synthesizer.TTSData) error {
	if !data.Input.AllowCollectData {
		goapp.Log.Info("Skip saving to DB")
		return nil
	}
	defer goapp.Estimate("SaveToDB")()

	return p.s.Save(data.RequestID, data.OriginalText, utils.RequestMain)
}
