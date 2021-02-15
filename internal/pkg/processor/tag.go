package processor

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type tagger struct {
	httpWrap HTTPInvoker
}

//NewTagger creates new processor
func NewTagger(urlStr string) (synthesizer.Processor, error) {
	res := &tagger{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *tagger) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip tagger")
		return nil
	}
	var output []*TaggedWord
	err := p.httpWrap.InvokeText(data.TextWithNumbers, &output)
	if err != nil {
		return err
	}
	data.Words = mapTagResult(output)
	return nil
}

//TaggedWord - tagger's result
type TaggedWord struct {
	Type   string
	String string
	Mi     string
	Lemma  string
}

func mapTagResult(tags []*TaggedWord) []*synthesizer.ProcessedWord {
	res := make([]*synthesizer.ProcessedWord, 0)
	for _, t := range tags {
		pw := synthesizer.ProcessedWord{Tagged: mapTag(t)}
		res = append(res, &pw)
	}
	return res
}

func mapTag(tag *TaggedWord) synthesizer.TaggedWord {
	res := synthesizer.TaggedWord{}
	if tag.Type == "SEPARATOR" {
		res.Separator = tag.String
	} else if tag.Type == "SENTENCE_END" {
		res.SentenceEnd = true
	} else if tag.Type == "WORD" {
		res.Word = tag.String
		res.Lemma = tag.Lemma
		res.Mi = tag.Mi
	} else if tag.Type == "SPACE" {
		res.Space = true
	}
	return res
}

func (p *tagger) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
