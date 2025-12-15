package processor

import (
	"context"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

type readSymbols struct {
	symbolToWords  map[string][]string
	defaultSymbols map[string][]string
}

func NewReadSymbols() (synthesizer.Processor, error) {
	sm := initSymbolsToWords()
	ds := initDefaultSymbolsToWords(sm)

	res := &readSymbols{symbolToWords: sm, defaultSymbols: ds}

	return res, nil
}

func initDefaultSymbolsToWords(sm map[string][]string) map[string][]string {
	res := make(map[string][]string)
	for _, s := range []string{"(", ")", "|", "%", "*", "+", "=", "{", "}", "$", "_", "[", "]"} {
		if words, ok := sm[s]; ok {
			res[s] = words
		}
	}
	return res
}

func initSymbolsToWords() map[string][]string {
	res := make(map[string][]string)
	res["("] = []string{"skliaustai", "atsidaro"}
	res[")"] = []string{"skliaustai", "užsidaro"}
	res["|"] = []string{"vertikalus", "brūkšnys"}
	res["%"] = []string{"procentas"}
	res["*"] = []string{"žvaigždutė"}
	res["+"] = []string{"pliusas"}
	res["="] = []string{"lygu"}
	res["{"] = []string{"figūrinių", "skliaustų", "atidarymas"}
	res["}"] = []string{"figūrinių", "skliaustų", "uždarymas"}
	res["["] = []string{"laužtinių", "skliaustų", "atidarymas"}
	res["]"] = []string{"laužtinių", "skliaustų", "uždarymas"}
	res["$"] = []string{"dolerio", "ženklas"}
	res["?"] = []string{"klaustukas"}
	res["!"] = []string{"šauktukas"}
	res[":"] = []string{"dvitaškis"}
	res[";"] = []string{"kabliataškis"}
	res[","] = []string{"kablelis"}
	res["."] = []string{"taškas"}
	res["-"] = []string{"brūkšnys"}
	res["_"] = []string{"pabraukimas"}
	return res
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
	mp := p.defaultSymbols
	if data.Cfg.Input.SymbolMode == api.SymbolModeReadSelected {
		mp = make(map[string][]string)
		for _, s := range data.Cfg.Input.SelectedSymbols {
			if words, ok := p.symbolToWords[s]; ok {
				mp[s] = words
			} else {
				log.Ctx(ctx).Warn().Str("symbol", s).Msg("Unknown symbol to read")
			}
		}
	}
	var err error
	data.Words, err = readSymbolsInWords(ctx, data.Words, mp)
	if err != nil {
		return err
	}
	return nil
}

func readSymbolsInWords(ctx context.Context, processedWord []*synthesizer.ProcessedWord, mp map[string][]string) ([]*synthesizer.ProcessedWord, error) {
	var res []*synthesizer.ProcessedWord
	for _, pw := range processedWord {
		if pw.Tagged.IsWord() {
			res = append(res, pw)
			continue
		}
		sep := pw.Tagged.Separator
		if words, ok := mp[sep]; ok {
			for _, w := range words {
				newPW := pw.Clone()
				newPW.Tagged.Word = w
				newPW.Tagged.Separator = ""
				res = append(res, newPW)
			}
		} else {
			res = append(res, pw)
		}

	}
	return res, nil
}

func (p *readSymbols) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

func (p *readSymbols) Info() string {
	return "readSymbols()"
}
