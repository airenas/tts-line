package processor

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/go-app/pkg/goapp"
	accentI "github.com/airenas/tts-line/internal/pkg/accent"
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
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*20)

	if err != nil {
		return nil, errors.Wrap(err, "can't init http client")
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
	} else if tag.Type == "WORD" || tag.Type == "NUMBER" {
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

// Info return info about processor
func (p *tagger) Info() string {
	return fmt.Sprintf("tagger(%s)", utils.RetrieveInfo(p.httpWrap))
}

type taggerAccents struct {
	httpWrap HTTPInvoker
}

//NewTaggerAccents creates new processor
func NewTaggerAccents(urlStr string) (synthesizer.Processor, error) {
	res := &taggerAccents{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*15)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *taggerAccents) Process(data *synthesizer.TTSData) error {
	var output []*TaggedWord
	data.TextWithNumbers = clearAccents(data.OriginalText)
	err := p.httpWrap.InvokeText(data.TextWithNumbers, &output)
	if err != nil {
		return err
	}
	data.Words, err = mapTagAccentResult(output, data.OriginalText)
	return err
}

func clearAccents(v string) string {
	rns := []rune(v)
	sb := strings.Builder{}
	for i := 0; i < len(rns); i++ {
		if i < (len(rns)-3) && rns[i] == '{' &&
			unicode.IsLetter(rns[i+1]) && accentI.Value(rns[i+2]) > 0 && rns[i+3] == '}' {
			sb.WriteRune(rns[i+1])
			i = i + 3
		} else {
			sb.WriteRune(rns[i])
		}
	}
	return sb.String()
}

func mapTagAccentResult(tags []*TaggedWord, text string) ([]*synthesizer.ProcessedWord, error) {
	res := make([]*synthesizer.ProcessedWord, 0)
	pos := 0
	rns := []rune(text)
	for _, t := range tags {
		acc, ws, err := moveText(rns, pos, t)
		if err != nil {
			return nil, errors.Wrapf(err, "Wrong data at %d, '%s' vs '%s'", pos, t.String, string(rns[pos:min(len(rns), pos+20)]))
		}
		pw := synthesizer.ProcessedWord{Tagged: mapTag(t)}
		if acc > 0 {
			pw.UserAccent = acc
		}
		pos = pos + ws
		res = append(res, &pw)
	}
	return res, nil
}

func moveText(rns []rune, pos int, tag *TaggedWord) (int, int, error) {
	if tag.Type == "SENTENCE_END" {
		return 0, 0, nil
	}
	if pos >= len(rns) {
		return 0, 0, errors.Errorf("Wrong position %d", pos)
	}
	tr := []rune(tag.String)
	if tag.Type == "WORD" {
		i := pos
		acc := 0
		for _, r := range tr {
			if i >= len(rns) {
				return 0, 0, errors.Errorf("Wrong word at '%s', wanted '%s'", string(rns[pos:min(pos+20, len(rns))]), tag.String)
			}
			if r == rns[i] {
				i++
				continue
			}
			if rns[i] == '{' && i < len(rns)-3 {
				at := accentI.Value(rns[i+2])
				if r != rns[i+1] || at == 0 || rns[i+3] != '}' {
					return 0, 0, errors.Errorf("Wrong word at '%s'", string(rns[pos:min(pos+20, len(rns))]))
				}
				if acc != 0 {
					return 0, 0, errors.Wrapf(
						utils.NewErrBadAccent([]string{string(rns[pos:min(pos+20, len(rns))])}),
						"Only one accent is allowed")
				}
				acc = at*100 + i - pos + 1
				i = i + 4
			} else {
				return 0, 0, errors.Errorf("Wrong word at '%s'", string(rns[pos:min(pos+20, len(rns))]))
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
		return nil, errors.Wrap(err, "Can't init http client")
	}
	return res, nil
}

func (p *ssmlTagger) Process(data *synthesizer.TTSData) error {
	var output []*TaggedWord
	txt := clearAccents(data.TextWithNumbers)
	err := p.httpWrap.InvokeText(txt, &output)
	if err != nil {
		return err
	}
	data.Words, err = mapTagAccentResult(output, data.TextWithNumbers)
	return err
}

// Info return info about processor
func (p *ssmlTagger) Info() string {
	return fmt.Sprintf("SSMLTagger(%s)", utils.RetrieveInfo(p.httpWrap))
}
