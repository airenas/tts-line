package processor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mvdan.cc/xurls/v2"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type urlFinder struct {
	urlRegexp  *regexp.Regexp
	emaiRegexp *regexp.Regexp
}

// NewURLReplacer creates new URL replacer processor
func NewURLFinder() (*urlFinder, error) {
	res := &urlFinder{}
	res.urlRegexp = xurls.Relaxed()
	// from https://html.spec.whatwg.org/#valid-e-mail-address
	res.emaiRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return res, nil
}

type urlReplacer struct {
	taggerHTTPWrap    HTTPInvokerJSON
	urlReaderHTTPWrap HTTPInvokerJSON
}

// NewURLReplacer creates new URL replacer processor
func NewURLReplacer(urlReaderURLStr, taggerURLStr string) (*urlReplacer, error) {
	res := &urlReplacer{}

	var err error
	res.urlReaderHTTPWrap, err = newHTTPWrapBackoff(urlReaderURLStr, time.Second*20)
	if err != nil {
		return nil, errors.Wrap(err, "can't init url reader http client")
	}
	log.Info().Str("url reader url", urlReaderURLStr).Msg("URL Replacer initialized")

	res.taggerHTTPWrap, err = newHTTPWrapBackoff(taggerURLStr, time.Second*20)
	if err != nil {
		return nil, errors.Wrap(err, "can't init tagger http client")
	}
	log.Info().Str("word tagger url", taggerURLStr).Msg("URL Replacer initialized")

	return res, nil
}

func (p *urlReplacer) Process(ctx context.Context, data *synthesizer.TTSData) error {
	if p.skip(data) {
		log.Ctx(ctx).Info().Msg("Skip url replace")
		return nil
	}
	defer goapp.Estimate("URL replace")()
	var err error
	data.Words, err = p.replaceURLs(ctx, data.Words)
	if err != nil {
		return fmt.Errorf("replace URLs: %w", err)
	}
	return nil
}

func (p *urlReplacer) replaceURLs(ctx context.Context, words []*synthesizer.ProcessedWord) ([]*synthesizer.ProcessedWord, error) {
	urls := collectURLs(words)
	if len(urls) == 0 {
		return words, nil
	}
	mappedURLs, err := p.mapURLs(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("map URLs: %w", err)
	}
	urlWords := collectWords(mappedURLs)
	taggedWords := make(map[string]*synthesizer.TaggedWord)
	if len(urlWords) != 0 {
		taggedWords, err = p.tagWords(ctx, urlWords)
		if err != nil {
			return nil, fmt.Errorf("tag words: %w", err)
		}
	}

	res := make([]*synthesizer.ProcessedWord, 0, len(words))
	for _, w := range words {
		if isURL(w) {
			mapped, ok := mappedURLs[w.Tagged.Word]
			if !ok {
				return nil, fmt.Errorf("missing mapped URL for %s", w.Tagged.Word)
			}
			for _, urlW := range mapped.expanded {
				if urlW.kind == urlPartWordTypeWord {
					tagged, ok := taggedWords[urlW.text]
					if !ok {
						return nil, fmt.Errorf("missing tagged words for %s", w.Tagged.Word)
					}
					res = append(res, &synthesizer.ProcessedWord{
						Tagged:   *tagged,
						TextPart: w.TextPart,
						FromWord: &w.Tagged})
				} else if urlW.kind == urlPartWordTypeChars {
					res = append(res, &synthesizer.ProcessedWord{
						Tagged: synthesizer.TaggedWord{
							Word:  urlW.text,
							Mi:    miAbbreviation,
							Lemma: urlW.text,
						},
						TextPart: w.TextPart,
						FromWord: &w.Tagged,
					})
				} else if urlW.kind == urlPartWordTypePunct {
					res = append(res, &synthesizer.ProcessedWord{
						Tagged: synthesizer.TaggedWord{
							Separator: urlW.text,
							Mi:        miComma,
						},
						TextPart: w.TextPart,
						FromWord: &w.Tagged,
					})
				} else {
					return nil, fmt.Errorf("unknown url part word type %s", urlW.kind)
				}
			}
		} else {
			res = append(res, w)
		}
	}
	return res, nil
}

func (p *urlReplacer) tagWords(ctx context.Context, urlWords map[string]struct{}) (map[string]*synthesizer.TaggedWord, error) {
	ctx, span := utils.StartSpan(ctx, "urlReplacer.tagWords")
	defer span.End()

	res := make(map[string]*synthesizer.TaggedWord)
	if len(urlWords) == 0 {
		return res, nil
	}
	input := [][]string{make([]string, 0, len(urlWords))}
	for w := range urlWords {
		input[0] = append(input[0], w)
	}
	var output []TaggedWord
	err := p.taggerHTTPWrap.InvokeJSON(ctx, input, &output)
	if err != nil {
		return nil, fmt.Errorf("invoke tagger: %w", err)
	}
	for _, w := range output {
		tw := mapTag(&w)
		if tw.SentenceEnd {
			continue
		}
		res[tw.Word] = &tw
	}
	return res, nil
}

func collectWords(mappedURLs map[string]*urlMap) map[string]struct{} {
	res := make(map[string]struct{})
	for _, v := range mappedURLs {
		for _, w := range v.expanded {
			if w.kind == urlPartWordTypeWord {
				res[w.text] = struct{}{}
			}
		}
	}
	return res
}

type expandedWord struct {
	text string
	kind urlPartWordType
}

type urlMap struct {
	orig     string
	expanded []*expandedWord
}

type urlReaderRequest struct {
	URLs []*urlReaderData `json:"items"`
}

type urlReaderDataType string

const (
	urlReaderDataTypeURL   urlReaderDataType = "url"
	urlReaderDataTypeEmail urlReaderDataType = "email"
)

type urlReaderData struct {
	Text string            `json:"text"`
	Type urlReaderDataType `json:"type"`
}

type urlReaderResponse struct {
	Items []*urlChangeData `json:"items"`
}

type urlChangeData struct {
	Text string         `json:"text"`
	Exp  []*urlPartWord `json:"expanded"`
}

type urlPartWordType string

const (
	urlPartWordTypeWord  urlPartWordType = "word"
	urlPartWordTypeChars urlPartWordType = "chars"
	urlPartWordTypePunct urlPartWordType = "punct"
)

type urlPartWord struct {
	Text string          `json:"text"`
	Kind urlPartWordType `json:"kind"`
}

func (p *urlReplacer) mapURLs(ctx context.Context, urls map[string]string) (map[string]*urlMap, error) {
	ctx, span := utils.StartSpan(ctx, "urlReplacer.mapURLs")
	defer span.End()

	input := &urlReaderRequest{URLs: make([]*urlReaderData, 0, len(urls))}
	for k, v := range urls {
		ud := &urlReaderData{Text: k}
		if v == miLink {
			ud.Type = urlReaderDataTypeURL
		} else if v == miEmail {
			ud.Type = urlReaderDataTypeEmail
		} else {
			return nil, fmt.Errorf("unknown url kind '%s' for url %s", v, k)
		}
		input.URLs = append(input.URLs, ud)
	}
	var output urlReaderResponse
	err := p.urlReaderHTTPWrap.InvokeJSON(ctx, input, &output)
	if err != nil {
		return nil, fmt.Errorf("invoke url reader: %w", err)
	}

	res := make(map[string]*urlMap)
	for _, v := range output.Items {
		r := mapURLResponse(v)
		res[r.orig] = r
	}
	return res, nil
}

func mapURLResponse(v *urlChangeData) *urlMap {
	res := &urlMap{orig: v.Text, expanded: make([]*expandedWord, 0, len(v.Exp))}
	for _, p := range v.Exp {
		r := &expandedWord{text: p.Text, kind: p.Kind}
		res.expanded = append(res.expanded, r)
	}
	return res
}

func isURL(word *synthesizer.ProcessedWord) bool {
	return word.Tagged.IsWord() && (word.Tagged.Mi == miLink || word.Tagged.Mi == miEmail)
}

func collectURLs(words []*synthesizer.ProcessedWord) map[string]string {
	res := make(map[string]string)
	for _, w := range words {
		if isURL(w) {
			res[w.Tagged.Word] = w.Tagged.Mi
		}
	}
	return res
}

func (p *urlFinder) replaceAll(text string, placeholder string) ([]string, string) {
	var links []string
	res := p.urlRegexp.ReplaceAllStringFunc(text, func(in string) string {
		links = append(links, in)
		return placeholder
	})
	return links, res
}

// baseURL removes http, https, www, and / at the end
func baseURL(s string) string {
	res := s
	for _, p := range [...]string{"https://", "http://", "www."} {
		if strings.HasPrefix(strings.ToLower(res), p) {
			res = res[len(p):]
		}
	}
	return strings.TrimSuffix(res, "/")
}

func (p *urlReplacer) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *urlReplacer) Info() string {
	return fmt.Sprintf("urlReplacer(%s, %s)", utils.RetrieveInfo(p.urlReaderHTTPWrap), utils.RetrieveInfo(p.taggerHTTPWrap))
}
