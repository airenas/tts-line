package processor

import (
	"regexp"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"mvdan.cc/xurls/v2"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type urlReplacer struct {
	phrase     string
	urlRegexp  *regexp.Regexp
	emaiRegexp *regexp.Regexp
	skipURLs   map[string]bool
}

//NewURLReplacer creates new URL replacer processor
func NewURLReplacer() synthesizer.Processor {
	res := &urlReplacer{}
	res.phrase = "Internetinis adresas"
	res.urlRegexp = xurls.Relaxed()
	// from https://html.spec.whatwg.org/#valid-e-mail-address
	res.emaiRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	res.skipURLs = map[string]bool{"lrt.lt": true, "vdu.lt": true, "lrs.lt": true}
	return res
}

func (p *urlReplacer) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip url replace")
		return nil
	}
	defer goapp.Estimate("URL replace")()
	utils.LogData("Input: ", data.CleanedText)
	data.Text = p.replaceURLs(data.CleanedText)
	utils.LogData("Output: ", data.Text)
	return nil
}

func (p *urlReplacer) replaceURLs(s string) string {
	return p.urlRegexp.ReplaceAllStringFunc(s, func(in string) string {
		// leave emails
		if p.emaiRegexp.MatchString(in) {
			return in
		}
		// leave some URL
		fixed := baseURL(in)
		if p.skipURLs[strings.ToLower(fixed)] {
			return fixed
		}
		return p.phrase
	})
}

//baseURL removes http, https, www, and / at the end
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