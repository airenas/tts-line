package processor

import (
	"context"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

type readSymbols struct {
}

func NewReadSymbols() (synthesizer.Processor, error) {

	res := &readSymbols{}

	return res, nil
}

func (p *readSymbols) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "readSymbols.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip readSymbols")
		return nil
	}
	if data.Cfg.Input.SymbolMode == api.SymbolModeNone {
		log.Ctx(ctx).Info().Msg("Skip readSymbols - mode none")
		return nil
	}
	var mp map[string]struct{}
	modeType := synthesizer.NERReadableSymbol
	if data.Cfg.Input.SymbolMode == api.SymbolModeReadSelected {
		modeType = synthesizer.NERReadableAllSymbol
		mp = make(map[string]struct{})
		for _, s := range data.Cfg.Input.SelectedSymbols {
			mp[s] = struct{}{}
		}
	}
	if err := markToReadSymbols(ctx, data.Words, mp, modeType); err != nil {
		return err
	}
	return nil
}

func markToReadSymbols(_ctx context.Context, processedWord []*synthesizer.ProcessedWord, mp map[string]struct{}, modeType synthesizer.NEREnum) error {
	for _, pw := range processedWord {
		if pw.Tagged.IsWord() {
			continue
		}
		sep := pw.Tagged.Separator
		if sep == "" {
			continue
		}
		ok := mp == nil
		if !ok {
			_, ok = mp[sep]
		}
		if ok {
			pw.NERType = modeType
		}
	}
	return nil
}

func (p *readSymbols) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

func (p *readSymbols) Info() string {
	return "readSymbols()"
}
