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

// NewValidator creates new processor
func NewValidator(defaultMaxLen int) (synthesizer.Processor, error) {
	res := &validator{}
	if defaultMaxLen < 100 {
		return nil, errors.Errorf("wrong max input len %d. (>= 100)", defaultMaxLen)
	}
	res.defaultMax = defaultMaxLen
	return res, nil
}

func (p *validator) Process(data *synthesizer.TTSData) error {
	if skip(data) {
		goapp.Log.Info("Skip validator")
		return nil
	}
	return validate(getLen(data.Input.Text), getMaxLen(p.defaultMax, data.Input.AllowedMaxLen))
}

func getLen(s string) int {
	return utf8.RuneCountInString(strings.TrimSpace(s))
}

func getMaxLen(def, req int) int {
	if req > 0 {
		return req
	}
	return def
}

func validate(l, maxLen int) error {
	if l == 0 {
		return utils.ErrNoInput
	}
	if l > maxLen {
		return utils.NewErrTextTooLong(l, maxLen)
	}
	return nil
}

func skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

type ssmlValidator struct {
	defaultMax int
}

// NewSSMLValidator creates new processor
func NewSSMLValidator(defaultMaxLen int) (synthesizer.Processor, error) {
	res := &ssmlValidator{}
	if defaultMaxLen < 100 {
		return nil, errors.Errorf("wrong max input len %d. (>= 100)", defaultMaxLen)
	}
	res.defaultMax = defaultMaxLen
	return res, nil
}

func (p *ssmlValidator) Process(data *synthesizer.TTSData) error {
	if skip(data) {
		goapp.Log.Info("Skip validator")
		return nil
	}
	return validate(getSSMLTextLen(data), getMaxLen(p.defaultMax, data.Input.AllowedMaxLen))
}

func getSSMLTextLen(data *synthesizer.TTSData) int {
	res := 0
	for _, p := range data.SSMLParts {
		res += getLen(p.OriginalText)
	} 
	return res
}

