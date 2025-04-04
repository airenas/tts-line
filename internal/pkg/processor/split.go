package processor

import (
	"context"
	"fmt"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type splitter struct {
	maxChars int
}

// NewSplitter split text into batches
func NewSplitter(maxChars int) synthesizer.Processor {
	if maxChars < 1 {
		maxChars = 400
	}
	return &splitter{maxChars: maxChars}
}

func (p *splitter) Process(ctx context.Context, data *synthesizer.TTSData) error {
	var err error
	if p.custom(data) {
		log.Ctx(ctx).Info().Msg("Custom split")
		data.Parts, err = splitCustom(data)
		if err != nil {
			return err
		}
	} else {
		data.Parts, err = split(data.Words, p.maxChars)
		if err != nil {
			return err
		}
	}
	for _, p := range data.Parts {
		p.Cfg = &data.Cfg
	}

	utils.LogData(ctx, "Output", fmt.Sprintf("split into %d", len(data.Parts)), nil)
	return nil
}

func split(data []*synthesizer.ProcessedWord, max int) ([]*synthesizer.TTSDataPart, error) {
	res := []*synthesizer.TTSDataPart{}
	from := 0
	l := len(data)
	for from < l {
		to := findSentenceEnd(data, from, max, func(tw, twPr *synthesizer.TaggedWord) bool { return tw.SentenceEnd })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from:to]})
			from = to
			continue
		}
		to = findSentenceEnd(data, from, max, func(tw, twPr *synthesizer.TaggedWord) bool {
			return twPr != nil && twPr.Separator != "" && tw.Space
		})
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from:to]})
			from = to
			continue
		}
		to = findSentenceEnd(data, from, max, func(tw, twPr *synthesizer.TaggedWord) bool { return tw.Separator != "" })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from:to]})
			from = to
			continue
		}
		to = findSentenceEnd(data, from, max, func(tw, twPr *synthesizer.TaggedWord) bool { return true })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from:to]})
			from = to
			continue
		} else {
			return nil, errors.Errorf("Can' split into no longer than %d sequences", max)
		}
	}
	if len(res) > 0 {
		res[0].First = true
	}
	return res, nil
}

func findSentenceEnd(data []*synthesizer.ProcessedWord, from int, max int, cmp func(tw, twPr *synthesizer.TaggedWord) bool) int {
	chCount := 0
	lastI := from - 1
	var pr *synthesizer.TaggedWord
	for i := from; i < len(data); i++ {
		tw := &data[i].Tagged
		if tw.Word != "" {
			chCount += len(tw.Word)
		}
		if chCount > max {
			return lastI + 1
		}
		if cmp(tw, pr) && chCount > 0 {
			lastI = i
		}
		pr = tw
	}
	return len(data)
}

func splitCustom(data *synthesizer.TTSData) ([]*synthesizer.TTSDataPart, error) {
	res := []*synthesizer.TTSDataPart{}
	res = append(res, &synthesizer.TTSDataPart{Text: data.OriginalText, First: true})
	return res, nil
}

func (p *splitter) custom(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
