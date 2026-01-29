package processor

import (
	"context"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

// LoadDB interface for text saving
type LoadDB interface {
	LoadText(req string, reqType utils.RequestTypeEnum) (string, error)
}

type loader struct {
	sDB LoadDB
}

// NewLoader creates new text to db processor
func NewLoader(s LoadDB) (synthesizer.Processor, error) {
	if s == nil {
		return nil, errors.New("No Saver")
	}
	return &loader{sDB: s}, nil
}

func (p *loader) Process(ctx context.Context, data *synthesizer.TTSData) error {
	defer goapp.Estimate("LoadDB")()

	var err error
	data.PreviousText, err = p.sDB.LoadText(data.Input.RequestID, utils.RequestNormalized)
	if err != nil {
		return errors.Wrapf(err, "Can't load request from DB for id '%s'", data.Input.RequestID)
	}
	return nil
}
