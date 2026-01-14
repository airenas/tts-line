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
	urlPhrase   string
	emailPhrase string
	// urlFinder *urlFinder
	skipURLs map[string]bool

	taggerHTTPWrap HTTPInvokerJSON
}

// NewURLReplacer creates new URL replacer processor
func NewURLReplacer(taggerURLStr string) (*urlReplacer, error) {
	res := &urlReplacer{}
	res.urlPhrase = "Internetinis adresas"
	res.emailPhrase = "Elektroninio pašto adresas"

	var err error
	res.taggerHTTPWrap, err = newHTTPWrapBackoff(taggerURLStr, time.Second*20)
	if err != nil {
		return nil, errors.Wrap(err, "can't init tagger http client")
	}
	log.Info().Str("word tagger url", taggerURLStr).Msg("URL Replacer initialized")

	res.skipURLs = map[string]bool{"lrt.lt": true, "vdu.lt": true, "lrs.lt": true}
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
	if len(urlWords) == 0 {
		return words, nil
	}
	taggedWords, err := p.tagWords(ctx, urlWords)
	if err != nil {
		return nil, fmt.Errorf("tag words: %w", err)
	}

	res := make([]*synthesizer.ProcessedWord, 0, len(words))
	for _, w := range words {
		if isURL(w) {
			mapped, ok := mappedURLs[w.Tagged.Word]
			if !ok {
				return nil, fmt.Errorf("missing mapped URL for %s", w.Tagged.Word)
			}
			for _, urlW := range mapped.expanded {
				tagged, ok := taggedWords[urlW]
				if !ok {
					return nil, fmt.Errorf("missing tagged words for %s", w.Tagged.Word)
				}
				res = append(res, &synthesizer.ProcessedWord{
					Tagged:   *tagged,
					TextPart: w.TextPart,
				})
			}
		} else {
			res = append(res, w)
		}
	}
	return res, nil
}

func (p *urlReplacer) tagWords(ctx context.Context, urlWords map[string]struct{}) (map[string]*synthesizer.TaggedWord, error) {
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
			res[w] = struct{}{}
		}
	}
	return res
}

type urlMap struct {
	orig     string
	expanded []string
}

func (p *urlReplacer) mapURLs(ctx context.Context, urls map[string]string) (map[string]*urlMap, error) {
	res := make(map[string]*urlMap)
	for k, v := range urls {
		if v == linkMI {
			l := baseURL(k)
			if p.skipURLs[l] {
				res[k] = &urlMap{orig: k, expanded: []string{l}}
			} else {
				res[k] = &urlMap{orig: k, expanded: strings.Split(p.urlPhrase, " ")}
			}
		} else if v == emailMI {
			res[k] = &urlMap{orig: k, expanded: strings.Split(p.emailPhrase, " ")}
		}
	}
	return res, nil
}

func isURL(word *synthesizer.ProcessedWord) bool {
	return word.Tagged.IsWord() && (word.Tagged.Mi == linkMI || word.Tagged.Mi == emailMI)
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
