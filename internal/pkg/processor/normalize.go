package processor

import (
	"regexp"
	"strings"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type normalizer struct {
}

//NewNormalizer creates new text normalizar processor
func NewNormalizer() synthesizer.Processor {
	return &normalizer{}
}

func (p *normalizer) Process(data *synthesizer.TTSData) error {
	utils.LogData("Input: ", data.OriginalText)
	data.Text = normalize(data.OriginalText)
	if data.Text == "" {
		data.ValidationFailures = []api.ValidateFailure{api.ValidateFailure{Check: api.Check{ID: "no_text"}}}
	}
	utils.LogData("Output: ", data.Text)
	return nil
}

var regSpaces = regexp.MustCompile(`[ ]{2,}`)

func normalize(text string) string {
	res := strings.ReplaceAll(text, "‘", "`")
	res = strings.ReplaceAll(res, "–", "-")
	res = strings.ReplaceAll(res, "‑", "-")
	res = strings.ReplaceAll(res, "–", "-")
	//res = strings.ReplaceAll(res, "-", " - ")
	return normalizeSpaces(res)
}

func normalizeSpaces(text string) string {
	res := strings.ReplaceAll(text, "\n", " ")
	res = strings.ReplaceAll(res, "\t", " ")
	res = strings.ReplaceAll(res, "\r", " ")
	res = strings.TrimSpace(regSpaces.ReplaceAllString(res, " "))
	return res
}
