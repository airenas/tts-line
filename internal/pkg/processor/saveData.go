package processor

import (
	"fmt"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

//SaverDB interface for text saving
type SaverDB interface {
	Save(req, text string, reqType utils.RequestTypeEnum, tags []string) error
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

	return p.sDB.Save(data.RequestID, getText(data, p.tType), p.tType, getTags(data.Input))
}

func getText(data *synthesizer.TTSData, t utils.RequestTypeEnum) string {
	switch t {
	case utils.RequestOriginal:
		return data.OriginalText
	case utils.RequestCleaned:
		return strings.Join(data.Text, " ")
	case utils.RequestNormalized:
		return strings.Join(data.TextWithNumbers, " ")
	case utils.RequestUser:
		return data.OriginalText
	case utils.RequestOriginalSSML:
		return data.OriginalText
	}
	goapp.Log.Warnf("Not configured RequestTypeEnum %v", t)
	return data.OriginalText
}

func getTags(cfg *api.TTSRequestConfig) []string {
	if cfg == nil {
		return nil
	}
	return cfg.SaveTags
}

// Info return info about processor
func (p *saver) Info() string {
	return fmt.Sprintf("saver(%s)", p.tType.String())
}
