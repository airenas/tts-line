package processor

import (
	"fmt"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type splitter struct {
}

//NewSplitter split text into batches
func NewSplitter() synthesizer.Processor {
	return &splitter{}
}

func (p *splitter) Process(data *synthesizer.TTSData) error {
	var err error
	data.Parts, err = split(data.Words)
	if err != nil {
		return err
	}
	utils.LogData("Output: ", fmt.Sprintf("split into %d", len(data.Parts)))
	return nil
}

func split(data []*synthesizer.ProcessedWord) ([]*synthesizer.TTSDataPart, error) {
	res := []*synthesizer.TTSDataPart{}
	from := 0
	l := len(data)
	max := 400
	for from < l {
		to := findSentenceEnd(data, from, max, func(tw *synthesizer.TaggedWord) bool { return tw.SentenceEnd })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from : to+1]})
			from = to + 1
			continue
		}
		to = findSentenceEnd(data, from, max, func(tw *synthesizer.TaggedWord) bool { return tw.Separator != "" })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from : to+1]})
			from = to + 1
			continue
		}
		to = findSentenceEnd(data, from, max, func(tw *synthesizer.TaggedWord) bool { return true })
		if to > from {
			res = append(res, &synthesizer.TTSDataPart{Words: data[from : to+1]})
			from = to + 1
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

func findSentenceEnd(data []*synthesizer.ProcessedWord, from int, max int, cmp func(tw *synthesizer.TaggedWord) bool) int {
	chCount := 0
	lastI := from
	for i := from; i < len(data); i++ {
		tw := &data[i].Tagged
		if cmp(tw) {
			lastI = i
		}
		if tw.Word != "" {
			chCount += len(tw.Word)
		}
		if chCount >= max {
			return lastI
		}
	}
	return len(data) - 1
}
