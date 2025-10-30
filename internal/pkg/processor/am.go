package processor

import (
	"context"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
)

type amodel struct {
	httpWrap    HTTPInvokerJSON
	url         string
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
		`"ioi`: `"io: i`,
		`"iOi`: `"io i`,
		`ioi`:  `io: i`,
		`iOi`:  `io i`,
	}
}

// NewAcousticModel creates new processor
func NewAcousticModel(config *viper.Viper) (synthesizer.PartProcessor, error) {
	if config == nil {
		return nil, errors.New("No acousticModel config")
	}

	res := &amodel{}
	res.url = config.GetString("url")
	am, err := utils.NewHTTPWrapT(getVoiceURL(res.url, "testVoice"), time.Second*120)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
	}
	res.httpWrap, err = utils.NewHTTPBackoff(am, newGPUBackoff, utils.RetryAll)
	if err != nil {
		return nil, errors.Wrap(err, "can't init AM client")
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
	log.Info().Msgf("AM pause: '%s', end symbol: '%s'", res.spaceSymbol, res.endSymbol)
	log.Info().Msgf("AM hasVocoder: %t", res.hasVocoder)
	return res, nil
}

func (p *amodel) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "amodel.Process")
	defer span.End()

	if data.Cfg.Input.OutputFormat == api.AudioNone {
		if data.Cfg.Input.OutputTextFormat == api.TextTranscribed {
			inData, _ := p.mapAMInput(data)
			data.TranscribedText = inData.Text
		}
		return nil
	}

	inData, inIndexes := p.mapAMInput(data)
	data.TranscribedText = inData.Text
	var output amOutput
	err := p.httpWrap.InvokeJSONU(ctx, getVoiceURL(p.url, data.Cfg.Voice), inData, &output)
	if err != nil {
		return err
	}
	// bug in am model
	fixSilDuration := 2
	data.DefaultSilence = output.SilDuration * fixSilDuration
	data.Step = output.Step
	data.Durations = output.Durations
	if err := mapAMOutputDurations(ctx, data, output.Durations, inIndexes); err != nil {
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
	Text     string  `json:"text"`
	Speed    float64 `json:"speedAlpha,omitempty"`
	Voice    string  `json:"voice"`
	Priority int     `json:"priority,omitempty"`
}

type amOutput struct {
	Data        string `json:"data"`
	Durations   []int  `json:"durations,omitempty"`
	SilDuration int    `json:"silDuration,omitempty"`
	Step        int    `json:"step,omitempty"`
}

func mapAMOutputDurations(ctx context.Context, data *synthesizer.TTSDataPart, durations []int, indRes []*synthesizer.SynthesizedPos) error {
	sums := make([]int, len(durations)+1)
	for i := 0; i < len(durations); i++ {
		sums[i+1] = sums[i] + durations[i]
	}
	for i, w := range data.Words {
		fromI := indRes[i].From
		toI := indRes[i].To
		if fromI < 0 || fromI >= len(sums) || toI < 0 || toI >= len(sums) {
			log.Ctx(ctx).Warn().Msgf("Invalid duration index %d %d %d %d", fromI, toI, len(sums), len(indRes))
			continue
		}
		w.SynthesizedPos = &synthesizer.SynthesizedPos{
			From: sums[fromI],
			To:   sums[toI],
		}
	}
	return nil
}

func (p *amodel) mapAMInput(data *synthesizer.TTSDataPart) (*amInput, []*synthesizer.SynthesizedPos) {
	res := &amInput{}
	res.Speed = data.Cfg.Speed
	res.Voice = data.Cfg.Voice
	res.Priority = data.Cfg.Input.Priority
	if data.Cfg.JustAM {
		res.Text = data.Text
		return res, nil
	}
	sb := make([]string, 0)
	//sb := &strings.Builder{}
	pause := p.spaceSymbol
	sb = append(sb, pause)
	lastSep := ""
	indRes := make([]*synthesizer.SynthesizedPos, len(data.Words))
	for i, w := range data.Words {
		indRes[i] = &synthesizer.SynthesizedPos{}
		indRes[i].From = len(sb)
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
		indRes[i].To = len(sb)
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
	if len(indRes) > 0 {
		lEnd := len(strings.Split(p.endSymbol, " "))
		if lEnd > 1 {
			indRes[len(indRes)-1].To += (lEnd - 1)
		}
	}
	res.Text = strings.Join(sb, " ")
	return res, indRes
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

func newGPUBackoff() backoff.BackOff {
	res := backoff.NewExponentialBackOff()
	res.InitialInterval = time.Second * 2
	return backoff.WithMaxRetries(res, 3)
}

func getVoiceURL(url, voice string) string {
	return strings.Replace(url, "{{voice}}", voice, -1)
}
