package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/airenas/tts-line/internal/pkg/accent"
	accentI "github.com/airenas/tts-line/internal/pkg/accent"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type tagger struct {
	httpWrap HTTPInvoker
}

// NewTagger creates new processor
func NewTagger(urlStr string) (synthesizer.Processor, error) {
	res := &tagger{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *tagger) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "tagger.Process")
	defer span.End()

	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip tagger")
		return nil
	}
	var output []*TaggedWord
	err := p.httpWrap.InvokeText(ctx, strings.Join(data.TextWithNumbers, ""), &output)
	if err != nil {
		return err
	}
	data.Words = mapTagResult(output)

	if !hasWords(data.Words) {
		return utils.ErrNoInput
	}
	return nil
}

func hasWords(processedWord []*synthesizer.ProcessedWord) bool {
	for _, w := range processedWord {
		if w.Tagged.IsWord() {
			return true
		}
	}
	return false
}

// TaggedWord - tagger's result
type TaggedWord struct {
	Type   string `json:"type"`
	String string `json:"string,omitempty"`
	Mi     string `json:"mi,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
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
	switch tag.Type {
	case "SEPARATOR":
		res.Separator = tag.String
		res.Mi = tag.Mi
	case "SENTENCE_END":
		res.SentenceEnd = true
	case "WORD", "NUMBER":
		res.Word = tag.String
		res.Lemma = tag.Lemma
		res.Mi = tag.Mi
	case "SPACE":
		res.Space = true
	}
	return res
}

func (p *tagger) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *tagger) Info() string {
	return fmt.Sprintf("tagger(%s)", utils.RetrieveInfo(p.httpWrap))
}

type taggerAccents struct {
	httpWrap HTTPInvoker
}

// NewTaggerAccents creates new processor
func NewTaggerAccents(urlStr string) (synthesizer.Processor, error) {
	res := &taggerAccents{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*15)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *taggerAccents) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "taggerAccents.Process")
	defer span.End()

	var output []*TaggedWord
	data.TextWithNumbers = []string{data.OriginalText}
	err := p.httpWrap.InvokeText(ctx, accent.ClearAccents(strings.Join(data.TextWithNumbers, " ")), &output)
	if err != nil {
		return err
	}
	data.Words, err = mapTagAccentResult(output, data.TextWithNumbers, data.OriginalTextParts)
	return err
}

func mapTagAccentResult(tags []*TaggedWord, text []string, textParts []*synthesizer.TTSTextPart) ([]*synthesizer.ProcessedWord, error) {
	res := make([]*synthesizer.ProcessedWord, 0)
	pos, posI := 0, 0
	if len(text) <= posI {
		return nil, fmt.Errorf("no text")
	}
	rns := []rune(text[posI])
	for _, t := range tags {
		acc, ws, err := moveText(rns, pos, t)
		if err != nil {
			return nil, errors.Wrapf(err, "wrong data at %d, '%s' vs '%s'", pos, t.String, string(rns[pos:min(len(rns), pos+20)]))
		}
		pw := synthesizer.ProcessedWord{Tagged: mapTag(t)}
		if acc > 0 {
			pw.UserAccent = acc
		}
		if len(textParts) > posI {
			tp := textParts[posI]
			pw.TextPart = tp
		}
		res = append(res, &pw)
		pos += ws
		if pos >= len(rns) {
			posI++
			pos = 0
			if len(text) > posI {
				rns = []rune(text[posI])
			} else {
				rns = nil
			}
		}
	}
	// group words that are required to be spelled out, but where split by tagger
	finalRes := make([]*synthesizer.ProcessedWord, 0)
	var lastTP *synthesizer.TTSTextPart
	for i, w := range res {
		if lastTP != nil && w.TextPart == lastTP {
			if i < len(res)-1 {
				nextW := res[i+1]
				if nextW.TextPart != lastTP && !w.Tagged.IsWord() {
					finalRes = append(finalRes, w)
				}
			}
			continue
		}
		finalRes = append(finalRes, w)
		if w.TextPart != nil && w.TextPart.InterpretAs == ssml.InterpretAsTypeCharacters {
			lastTP = w.TextPart
			w.Tagged.Word = w.TextPart.Text
			w.Tagged.Separator = ""
		} else {
			lastTP = nil
		}
	}

	return finalRes, nil
}

func moveText(rns []rune, pos int, tag *TaggedWord) (int, int, error) {
	if tag.Type == "SENTENCE_END" {
		return 0, 0, nil
	}
	if pos >= len(rns) {
		return 0, 0, errors.Errorf("wrong position %d", pos)
	}
	tr := []rune(tag.String)
	if tag.Type == "WORD" {
		i := pos
		acc := 0
		for _, r := range tr {
			if i >= len(rns) {
				return 0, 0, errors.Errorf("wrong word at '%s', wanted '%s'", string(rns[pos:min(pos+20, len(rns))]), tag.String)
			}
			if r == rns[i] {
				i++
				continue
			}
			if rns[i] == '{' && i < len(rns)-3 {
				at := accentI.Value(rns[i+2])
				if r != rns[i+1] || at == 0 || rns[i+3] != '}' {
					return 0, 0, errors.Errorf("wrong word at '%s'", string(rns[pos:min(pos+20, len(rns))]))
				}
				if acc != 0 {
					return 0, 0, errors.Wrapf(
						utils.NewErrBadAccent([]string{string(rns[pos:min(pos+20, len(rns))])}),
						"only one accent is allowed")
				}
				acc = at*100 + i - pos + 1
				i = i + 4
			} else {
				return 0, 0, errors.Errorf("wrong word at '%s'", string(rns[pos:min(pos+20, len(rns))]))
			}
		}
		return acc, i - pos, nil
	}
	return 0, len(tr), nil
}

func min(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

// Info return info about processor
func (p *taggerAccents) Info() string {
	return fmt.Sprintf("taggerAccents(%s)", utils.RetrieveInfo(p.httpWrap))
}

type ssmlTagger struct {
	httpWrap HTTPInvoker
}

// NewSSMLTagger creates new processor
func NewSSMLTagger(urlStr string) (synthesizer.Processor, error) {
	res := &ssmlTagger{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*15)
	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
	}
	return res, nil
}

func (p *ssmlTagger) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "ssmlTagger.Process")
	defer span.End()

	var output []*TaggedWord
	data.TextWithNumbers = addSpaces(data.TextWithNumbers)
	txt := accent.ClearAccents(strings.Join(data.TextWithNumbers, " "))
	err := p.httpWrap.InvokeText(ctx, txt, &output)
	if err != nil {
		return err
	}
	data.Words, err = mapTagAccentResult(output, data.TextWithNumbers, data.OriginalTextParts)
	return err
}

func addSpaces(in []string) []string {
	res := make([]string, len(in))
	for i, s := range in {
		if strings.HasPrefix(s, " ") {
			res[i] = s
		} else {
			res[i] = s + " "
		}
	}
	return res
}

// Info return info about processor
func (p *ssmlTagger) Info() string {
	return fmt.Sprintf("SSMLTagger(%s)", utils.RetrieveInfo(p.httpWrap))
}
