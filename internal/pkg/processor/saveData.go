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
	sDB   SaverDB
	tType utils.RequestTypeEnum
}

//NewSaver creates new text to db processor
func NewSaver(s SaverDB, t utils.RequestTypeEnum) (synthesizer.Processor, error) {
	if s == nil {
		return nil, errors.New("No Saver")
	}
	return &saver{sDB: s, tType: t}, nil
}

func (p *saver) Process(data *synthesizer.TTSData) error {
	if !data.Input.AllowCollectData {
		goapp.Log.Info("Skip saving to DB")
		return nil
	}
	defer goapp.Estimate("SaveToDB " + p.tType.String())()

	return p.sDB.Save(data.RequestID, getText(data, p.tType), p.tType)
}

func getText(data *synthesizer.TTSData, t utils.RequestTypeEnum) string {
	if t == utils.RequestOriginal {
		return data.OriginalText
	}
	if t == utils.RequestCleaned {
		return data.Text
	}
	return data.TextWithNumbers
}
