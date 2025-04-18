package processor

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"

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

func (p *validator) Process(ctx context.Context, data *synthesizer.TTSData) error {
	if skip(data) {
		log.Ctx(ctx).Info().Msg("Skip validator")
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

func (p *ssmlValidator) Process(ctx context.Context, data *synthesizer.TTSData) error {
	if skip(data) {
		log.Ctx(ctx).Info().Msg("Skip validator")
		return nil
	}
	return validate(getSSMLTextLen(data), getMaxLen(p.defaultMax, data.Input.AllowedMaxLen))
}

func getSSMLTextLen(data *synthesizer.TTSData) int {
	res := 0
	for _, p := range data.SSMLParts {
		for _, tp := range p.OriginalTextParts {
			res += getLen(tp.Text)
		}
	}
	return res
}

// Info return info about processor
func (p *ssmlValidator) Info() string {
	return fmt.Sprintf("SSMLValidator(%d)", p.defaultMax)
}
