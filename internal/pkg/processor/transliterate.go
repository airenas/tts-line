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
	var res []*synthesizer.ProcessedWord
	for _, s := range sentences {
		sentenceRes, err := processSentence(ctx, p.httpWrap, s)
		if err != nil {
			return err
		}
		res = append(res, sentenceRes...)
	}
	data.Words = res
	return nil
}

func processSentence(ctx context.Context, client HTTPInvokerJSON, s []*synthesizer.ProcessedWord) ([]*synthesizer.ProcessedWord, error) {
	ctx, span := utils.StartSpan(ctx, "transliterator.processSentence")
	defer span.End()

	var input []*transliteratorInput
	for _, w := range s {
		input = append(input, toTransliteratorInput(w))
	}
	var output []*transliteratorOutput
	err := client.InvokeJSON(ctx, input, &output)
	if err != nil {
		return nil, err
	}
	return mapTransliteratorRes(output, s)
}

func toTransliteratorInput(w *synthesizer.ProcessedWord) *transliteratorInput {
	tw := w.Tagged
	if tw.Space {
		return &transliteratorInput{Type: tw.TypeStr(), String: " "}
	}
	if tw.Separator != "" {
		return &transliteratorInput{Type: tw.TypeStr(), String: tw.Separator}
	}
	lang := extractLanguage(w)
	return &transliteratorInput{Type: tw.TypeStr(), String: tw.Word, Mi: tw.Mi, Lemma: tw.Lemma, Language: lang}
}

func extractLanguage(word *synthesizer.ProcessedWord) string {
	if from := word.FromWord; from != nil {
		if from.Mi == miEmail || from.Mi == miLink {
			return langURL
		}
	}

	if part := word.TextPart; part != nil {
		return mapLanguage(part.Language)
	}
	return ""
}

func mapLanguage(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func mapTransliteratorRes(output []*transliteratorOutput, s []*synthesizer.ProcessedWord) ([]*synthesizer.ProcessedWord, error) {
	if len(output) != len(s) {
		return nil, fmt.Errorf("different length of input and output: %d != %d", len(s), len(output))
	}

	res := make([]*synthesizer.ProcessedWord, 0, len(s))
	for i, o := range output {
		w := s[i]
		tw := w.Tagged
		if tw.IsWord() {
			if !compareWords(tw.Word, o.String) {
				return nil, fmt.Errorf("different word: %s != %s", tw.Word, o.String)
			}
			if w.UserTranscription != "" || w.UserAccent != 0 { // do not overwrite user transcription
				res = append(res, w)
				continue
			}
			if len(o.Changes) == 0 {
				res = append(res, w)
				continue
			}

			for _, change := range o.Changes {
				if change.Type == transliteratorTypeWord {
					d := transcription.Parse(change.User)
					nw := &synthesizer.ProcessedWord{
						TextPart:          w.TextPart,
						UserTranscription: d.Transcription,
						UserSyllables:     d.Sylls,
						TranscriptionWord: d.Word,
					}
					nw.Tagged.Word = d.Word
					nw.Tagged.Mi = tw.Mi
					nw.Tagged.Lemma = d.Word
					res = append(res, nw)
				} else if change.Type == transliteratorTypePunct {
					res = append(res, &synthesizer.ProcessedWord{
						Tagged: synthesizer.TaggedWord{
							Separator: change.User,
							Mi:        miComma,
						},
						TextPart: w.TextPart,
					})
				}
			}
		} else {
			res = append(res, w)
		}
	}
	return res, nil
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
	Type     string `json:"type"`
	String   string `json:"string,omitempty"`
	Mi       string `json:"mi,omitempty"`
	Lemma    string `json:"lemma,omitempty"`
	Language string `json:"lang,omitempty"`
}

const langURL = "url" // fake language for transliteration

type transliteratorOutput struct {
	Type    string                        `json:"type"`
	String  string                        `json:"string,omitempty"`
	Mi      string                        `json:"mi,omitempty"`
	Lemma   string                        `json:"lemma,omitempty"`
	Changes []*transliteratorOutputChange `json:"change,omitempty"`
}

type transliteratorType string

const (
	transliteratorTypeWord  transliteratorType = "WORD"
	transliteratorTypeSpace transliteratorType = "SPACE"
	transliteratorTypePunct transliteratorType = "PUNCT"
)

type transliteratorOutputChange struct {
	Type transliteratorType `json:"type"`
	User string             `json:"user,omitempty"`
}
