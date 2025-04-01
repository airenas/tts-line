package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/transcription"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/rs/zerolog/log"
)

type transliterator struct {
	httpWrap HTTPInvokerJSON
}

func NewTransliterator(urlStr string) (synthesizer.Processor, error) {
	res := &transliterator{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, fmt.Errorf("init http client: %w", err)
	}
	return res, nil
}

func (p *transliterator) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "transliterator.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip transliterator")
		return nil
	}

	sentences, err := prepareSentences(data.Words)
	if err != nil {
		return err
	}
	log.Ctx(ctx).Debug().Int("len", len(sentences)).Msg("sentences")
	for _, s := range sentences {
		err := processSentence(ctx, p.httpWrap, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func processSentence(ctx context.Context, client HTTPInvokerJSON, s []*synthesizer.ProcessedWord) error {
	ctx, span := utils.StartSpan(ctx, "transliterator.processSentence")
	defer span.End()

	var input []*transliteratorInput
	for _, w := range s {
		tw := w.Tagged
		input = append(input, toTransliteratorInput(&tw))
	}
	var output []*transliteratorOutput
	err := client.InvokeJSON(ctx, input, &output)
	if err != nil {
		return err
	}
	return mapTransliteratorRes(output, s)
}

func toTransliteratorInput(taggedWord *synthesizer.TaggedWord) *transliteratorInput {
	if taggedWord.Space {
		return &transliteratorInput{Type: taggedWord.TypeStr(), String: " "}
	}
	if taggedWord.Separator != "" {
		return &transliteratorInput{Type: taggedWord.TypeStr(), String: taggedWord.Separator}
	}
	return &transliteratorInput{Type: taggedWord.TypeStr(), String: taggedWord.Word, Mi: taggedWord.Mi, Lemma: taggedWord.Lemma}
}

func mapTransliteratorRes(output []*transliteratorOutput, s []*synthesizer.ProcessedWord) error {
	if len(output) != len(s) {
		return fmt.Errorf("different length of input and output: %d != %d", len(s), len(output))
	}
	for i, o := range output {
		w := s[i]
		tw := w.Tagged
		if tw.IsWord() {
			if !compareWords(tw.Word, o.String) {
				return fmt.Errorf("different word: %s != %s", tw.Word, o.String)
			}
			if o.User != "" {
				if w.UserTranscription == "" && w.UserAccent == 0 { // do not overwrite user transcription
					d := transcription.Parse(o.User)
					w.UserTranscription = d.Transcription
					w.UserSyllables = d.Sylls
					w.TranscriptionWord = d.Word
				}
			}
		}
	}
	return nil
}

func compareWords(old, new string) bool {
	if old == new {
		return true
	}
	// allow "’" change
	if old == strings.ReplaceAll(new, "'", "’") {
		return true
	}
	// allow "‘" change
	if old == strings.ReplaceAll(new, "'", "`") {
		return true
	}
	return false
}

func prepareSentences(processedWord []*synthesizer.ProcessedWord) ([][]*synthesizer.ProcessedWord, error) {
	var res [][]*synthesizer.ProcessedWord
	var sentence []*synthesizer.ProcessedWord
	for _, w := range processedWord {
		sentence = append(sentence, w)
		if w.Tagged.SentenceEnd {
			res = append(res, sentence)
			sentence = []*synthesizer.ProcessedWord{}
		}
	}
	if len(sentence) > 0 {
		res = append(res, sentence)
	}
	return res, nil
}

func (p *transliterator) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

func (p *transliterator) Info() string {
	return fmt.Sprintf("transliterator(%s)", utils.RetrieveInfo(p.httpWrap))
}

// --------------------------------------------
type transliteratorInput struct {
	Type   string `json:"type"`
	String string `json:"string,omitempty"`
	Mi     string `json:"mi,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
}

type transliteratorOutput struct {
	Type   string `json:"type"`
	String string `json:"string,omitempty"`
	Mi     string `json:"mi,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
	User   string `json:"user,omitempty"`
}
