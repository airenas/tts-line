package processor

import (
	"strings"
	"unicode/utf8"

	"github.com/airenas/tts-line/internal/pkg/utils"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
)

type validator struct {
	defaultMax int
}

//NewValidator creates new processor
func NewValidator(defaultMaxLen int) (synthesizer.Processor, error) {
	res := &validator{}
	if defaultMaxLen < 100 {
		return nil, errors.Errorf("wrong max input len %d. (>= 100)", defaultMaxLen)
	}
	res.defaultMax = defaultMaxLen
	return res, nil
}

func (p *validator) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip validator")
		return nil
	}
	return validate(data.Input.Text, getMaxLen(p.defaultMax, data.Input.AllowedMaxLen))
}

func getMaxLen(def, req int) int {
	if req > 0 {
		return req
	}
	return def
}

func validate(text string, maxLen int) error {
	if strings.TrimSpace(text) == "" {
		return utils.ErrNoInput
	}
	if l := utf8.RuneCountInString(text); l > maxLen {
		return utils.NewErrTextTooLong(l, maxLen)
	}
	return nil
}

func (p *validator) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
