package processor

import (
	"regexp"

	"github.com/airenas/go-app/pkg/goapp"
	"mvdan.cc/xurls/v2"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type urlReplacer struct {
	phrase    string
	urlRegexp *regexp.Regexp
}

//NewURLReplacer creates new URL replacer processor
func NewURLReplacer() synthesizer.Processor {
	res := &urlReplacer{}
	res.phrase = "Internetinis adresas"
	res.urlRegexp = xurls.Relaxed()
	return res
}

func (p *urlReplacer) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip url replace")
		return nil
	}
	defer goapp.Estimate("URL replace")()
	utils.LogData("Input: ", data.CleanedText)
	data.Text = replaceURLs(data.CleanedText, p.urlRegexp, p.phrase)
	utils.LogData("Output: ", data.Text)
	return nil
}

func replaceURLs(s string, urlRegexp *regexp.Regexp, phrase string) string {
	return urlRegexp.ReplaceAllString(s, phrase)
}

func (p *urlReplacer) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}
