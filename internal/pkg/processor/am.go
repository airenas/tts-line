package processor

import (
	"strings"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type amodel struct {
	httpWrap    HTTPInvokerJSON
	spaceSymbol string
}

//NewAcousticModel creates new processor
func NewAcousticModel(urlStr string, spaceSym string) (synthesizer.PartProcessor, error) {
	res := &amodel{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	res.spaceSymbol = spaceSym
	if res.spaceSymbol == "" {
		res.spaceSymbol = "sil"
	}
	return res, nil
}

func (p *amodel) Process(data *synthesizer.TTSDataPart) error {
	inData := p.mapAMInput(data)
	var output amOutput
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}
	data.Spectogram = output.Data
	return nil
}

type amInput struct {
	Text string `json:"text"`
}

type amOutput struct {
	Data string `json:"data"`
}

func (p *amodel) mapAMInput(data *synthesizer.TTSDataPart) *amInput {
	res := &amInput{}
	var sb strings.Builder
	space := " "
	pause := p.spaceSymbol
	if data.First {
		sb.WriteString(pause)
	}
	lastSep := ""
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator != "" {
			sep := getSep(tgw.Separator)
			if sep != "" {
				sb.WriteString(space + sep)
				lastSep = sep
			}
			if addPause(sep) {
				sb.WriteString(space + pause)
			}
		} else if tgw.SentenceEnd {
			if getSep(lastSep) == "" {
				lastSep = "."
				sb.WriteString(space + lastSep)
			}
			if !strings.HasSuffix(sb.String(), pause) {
				sb.WriteString(space + pause)
			}
		} else {
			phns := strings.Split(w.Transcription, " ")
			for _, p := range phns {
				if !skipPhn(p) {
					sb.WriteString(space + p)
					lastSep = p
				}
			}
		}
	}
	if !strings.HasSuffix(sb.String(), pause) {
		sb.WriteString(space + pause)
	}
	res.Text = sb.String()
	return res
}

func getSep(s string) string {
	for _, sep := range []string{",", ".", "!", "?", "...", ":", "-"} {
		if s == sep {
			return s
		}
	}
	if s == ";" {
		return ","
	}
	return ""
}

func addPause(s string) bool {
	for _, sep := range []string{".", "!", "?", "...", ":"} {
		if s == sep {
			return true
		}
	}
	return false
}

func skipPhn(s string) bool {
	return s == "-"
}
