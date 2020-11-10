package processor

import (
	"regexp"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

type normalizer struct {
}

//NewNormalizer creates new text normalizar processor
func NewNormalizer() synthesizer.Processor {
	return &normalizer{}
}

func (p *normalizer) Process(data *synthesizer.TTSData) error {
	goapp.Log.Debugf("In: '%s'", data.OriginalText)
	data.Text = normalize(data.OriginalText)
	goapp.Log.Debugf("Out: '%s'", data.Text)
	return nil
}

var regSpaces = regexp.MustCompile(`[ ]{2,}`)

func normalize(text string) string {
	res := strings.ReplaceAll(text, "‘", "`")
	res = strings.ReplaceAll(res, "–", "-")
	res = strings.ReplaceAll(res, "‑", "-")
	res = strings.ReplaceAll(res, "–", "-")
	res = strings.ReplaceAll(res, "-", " - ")
	return normalizeSpaces(res)
}

func normalizeSpaces(text string) string {
	res := strings.ReplaceAll(text, "\n", " ")
	res = strings.ReplaceAll(res, "\t", " ")
	res = strings.ReplaceAll(res, "\r", " ")
	res = strings.TrimSpace(regSpaces.ReplaceAllString(res, " "))
	return res
}
