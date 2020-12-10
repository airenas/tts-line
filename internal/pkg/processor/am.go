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
	sb := &strings.Builder{}
	pause := p.spaceSymbol
	if data.First {
		write(sb, pause)
	}
	lastSep := ""
	for _, w := range data.Words {
		tgw := w.Tagged
		if tgw.Separator != "" {
			sep := getSep(tgw.Separator)
			if sep != "" {
				write(sb, sep)
				lastSep = sep
			}
			if addPause(sep) {
				write(sb, pause)
			}
		} else if tgw.SentenceEnd {
			if getSep(lastSep) == "" {
				lastSep = "."
				write(sb, lastSep)
			}
			if !strings.HasSuffix(sb.String(), pause) {
				write(sb, pause)
			}
		} else {
			phns := strings.Split(w.Transcription, " ")
			for _, p := range phns {
				if !skipPhn(p) {
					write(sb, p)
					lastSep = p
				}
			}
		}
	}
	if !strings.HasSuffix(sb.String(), pause) {
		write(sb, pause)
	}
	res.Text = sb.String()
	return res
}

func write(sb *strings.Builder, str string) {
	if sb.Len() > 0 {
		sb.WriteString(" " + str)
	} else {
		sb.WriteString(str)
	}
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
