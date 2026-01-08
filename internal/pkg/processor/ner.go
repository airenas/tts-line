package processor

import (
	"context"
	"unicode"

	"github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

// Does minimal NER processing
// - determines standalone letters
type ner struct {
}

func NewNER() (synthesizer.Processor, error) {
	res := &ner{}
	return res, nil
}

func (p *ner) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "ner.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip ner")
		return nil
	}

	sentences, err := prepareSentences(data.Words)
	if err != nil {
		return err
	}
	log.Ctx(ctx).Debug().Int("len", len(sentences)).Msg("sentences")
	for _, s := range sentences {
		err := processNERSentence(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func processNERSentence(_ctx context.Context, s []*synthesizer.ProcessedWord) error {
	i := 0
	for _, word := range s {
		if !word.Tagged.IsWord() {
			continue
		}
		w := accent.ClearAccents(word.Tagged.Word)
		runes := []rune(w)
		if i > 0 && len(runes) == 1 && unicode.IsLetter(runes[0]) && unicode.IsUpper(runes[0]) {
			word.NERType = synthesizer.NERSingleLetter
		} else if allGreekLetters(runes) {
			word.NERType = synthesizer.NERGreekLetters
		}

		i++
	}
	return nil
}

func allGreekLetters(runes []rune) bool {
	for _, r := range runes {
		if !(r >= 'Α' && r <= 'Ω') && !(r >= 'α' && r <= 'ω') {
			return false
		}
	}
	return len(runes) > 0
}

func (p *ner) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

func (p *ner) Info() string {
	return "ner()"
}
