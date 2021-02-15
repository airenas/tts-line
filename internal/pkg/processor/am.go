package processor

import (
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/spf13/viper"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type amodel struct {
	httpWrap    HTTPInvokerJSON
	spaceSymbol string
	endSymbol   string
	hasVocoder  bool
}

var trMap map[string]string

func init() {
	trMap = map[string]string{"\"Eu": "eu",
		"\"oi": "\"o: i", `"Oi`: `"o i`,
		`^Oi`: `"o i`, `^oi`: `"o i`,
		`Oi`: `o i`, `oi`: `o i`,
		`"ou`: `"o: u`, `"Ou`: `"o u`, `^ou`: `"o u`, `^Ou`: `"o u`,
		`ou`: `o u`, `Ou`: `o u`,
		`"iui`: `"iu i`,
	}
}

//NewAcousticModel creates new processor
func NewAcousticModel(config *viper.Viper) (synthesizer.PartProcessor, error) {
	if config == nil {
		return nil, errors.New("No acousticModel config")
	}

	res := &amodel{}
	var err error
	res.httpWrap, err = utils.NewHTTWrap(config.GetString("url"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't init http client")
	}
	res.spaceSymbol = config.GetString("spaceSymbol")
	if res.spaceSymbol == "" {
		res.spaceSymbol = "sil"
	}
	res.endSymbol = config.GetString("endSymbol")
	if res.endSymbol == "" {
		res.endSymbol = res.spaceSymbol
	}
	res.hasVocoder = config.GetBool("hasVocoder")
	goapp.Log.Infof("AM pause: '%s', end symbol: '%s'", res.spaceSymbol, res.endSymbol)
	goapp.Log.Infof("AM hasVocoder: %t", res.hasVocoder)
	return res, nil
}

func (p *amodel) Process(data *synthesizer.TTSDataPart) error {
	inData := p.mapAMInput(data)
	var output amOutput
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return err
	}
	if p.hasVocoder {
		data.Audio = output.Data
	} else {
		data.Spectogram = output.Data
	}
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
	if data.Cfg.JustAM {
		res.Text = data.Text
		return res
	}
	sb := make([]string, 0)
	//sb := &strings.Builder{}
	pause := p.spaceSymbol
	if data.First {
		sb = append(sb, pause)
	}
	lastSep := ""
	for i, w := range data.Words {
		tgw := w.Tagged
		if tgw.Space {
		} else if tgw.Separator != "" {
			sep := getSep(tgw.Separator, data.Words, i)
			if sep != "" {
				sb = append(sb, sep)
				lastSep = sep
			}
			if addPause(sep, data.Words, i) {
				sb = append(sb, pause)
			}
		} else if tgw.SentenceEnd {
			if lastSep == "" {
				lastSep = "."
				sb = append(sb, lastSep)
			}
			l := len(sb)
			if l > 0 && sb[l-1] != pause {
				sb = append(sb, pause)
			}
		} else {
			phns := strings.Split(w.Transcription, " ")
			for _, p := range phns {
				pt := changePhn(p)
				if pt != "" {
					sb = append(sb, pt)
					lastSep = ""
				}
			}
		}
	}

	l := len(sb)
	if l > 0 {
		if sb[l-1] != p.endSymbol {
			if sb[l-1] == pause {
				sb[l-1] = p.endSymbol
			} else {
				sb = append(sb, p.endSymbol)
			}
		}
	}
	res.Text = strings.Join(sb, " ")
	return res
}

func getSep(s string, words []*synthesizer.ProcessedWord, pos int) string {
	for _, sep := range [...]string{",", ".", "!", "?", "..."} {
		if s == sep {
			return s
		}
	}
	if s == ";" {
		return ","
	}
	if s == "-" || s == ":" {
		if pos > 0 && pos < len(words)-1 && words[pos-1].Tagged.IsWord() && words[pos+1].Tagged.IsWord() {
			return "" // drop dash and colon if between words
		}
		return s
	}
	return ""
}

func addPause(s string, words []*synthesizer.ProcessedWord, pos int) bool {
	for _, sep := range [...]string{".", "!", "?", "..."} {
		if s == sep {
			return true
		}
	}
	if s == "-" || s == ":" {
		if pos > 0 && pos < len(words)-1 {
			return !words[pos-1].Tagged.IsWord() || !words[pos+1].Tagged.IsWord()
		}
		return true
	}
	return false
}

func changePhn(s string) string {
	if s == "-" {
		return ""
	}
	ch := trMap[s]
	if ch != "" {
		return ch
	}
	return s
}
