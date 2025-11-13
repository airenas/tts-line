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
	"github.com/airenas/tts-line/internal/pkg/syntmodel"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/pkg/ssml"
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
	am = am.WithOutputFormat(utils.EncodingFormatMsgPack)
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
			inData, _ := p.mapAMInput(ctx, data)
			data.TranscribedText = inData.Text
		}
		return nil
	}

	inData, inIndexes := p.mapAMInput(ctx, data)
	data.TranscribedText = inData.Text
	var output syntmodel.AMOutput
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

func (p *amodel) mapAMInput(ctx context.Context, data *synthesizer.TTSDataPart) (*syntmodel.AMInput, []*synthesizer.SynthesizedPos) {
	res := &syntmodel.AMInput{}
	// res.Speed = calculateSpeed(data.Cfg.Prosodies)
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
	// split the last symbol if it has spaces
	if len(sb) > 0 {
		l := len(sb) - 1
		v := sb[l]
		vs := strings.Split(v, " ")
		if len(vs) > 1 {
			sb = sb[:l]
			sb = append(sb, vs...)
		}
	}
	res.Text = strings.Join(sb, " ")
	res.DurationsChange = prepareDurationsChange(ctx, sb, data.Cfg.Prosodies)
	res.PitchChange = preparePitchChange(ctx, sb, data.Cfg.Prosodies)
	return res, indRes
}

func preparePitchChange(ctx context.Context, sb []string, prosody []*ssml.Prosody) [][]*syntmodel.PitchChange {
	lastIsEmphasy := isEmphasyLast(prosody)
	isOneAccent, accentPos := isOneAccent(sb)
	if lastIsEmphasy && isOneAccent {
		// todo do separately
		log.Ctx(ctx).Debug().Int("pos", accentPos).Bool("emphasy", lastIsEmphasy).Msg("One accent with emphasy")
	}

	pc := makePitchChange(prosody)
	res := make([][]*syntmodel.PitchChange, len(sb))
	for i := 0; i < len(sb); i++ {
		res[i] = pc
	}
	return res
}

func makePitchChange(prosody []*ssml.Prosody) []*syntmodel.PitchChange {
	res := make([]*syntmodel.PitchChange, 0, len(prosody))
	for _, p := range prosody {
		switch p.Emphasis {
		case ssml.EmphasisTypeModerate:
			pc := &syntmodel.PitchChange{
				Value: 1.1,
				Type:  2,
			}
			res = append(res, pc)
		case ssml.EmphasisTypeStrong:
			pc := &syntmodel.PitchChange{
				Value: 1.2,
				Type:  2,
			}
			res = append(res, pc)
		case ssml.EmphasisTypeReduced:
			pc := &syntmodel.PitchChange{
				Value: 1 / 1.1,
				Type:  2,
			}
			res = append(res, pc)
		default:
			// no emphasis
		}

		switch p.Pitch.Kind {
		case ssml.PitchChangeNone:
			continue
		case ssml.PitchChangeHertz:
			pc := &syntmodel.PitchChange{
				Value: p.Pitch.Value,
				Type:  1,
			}
			res = append(res, pc)
		case ssml.PitchChangeMultiplier:
			pc := &syntmodel.PitchChange{
				Value: p.Pitch.Value,
				Type:  2,
			}
			res = append(res, pc)
		case ssml.PitchChangeSemitone:
			pc := &syntmodel.PitchChange{
				Value: p.Pitch.Value,
				Type:  3,
			}
			res = append(res, pc)
		}
	}
	return res
}

func prepareDurationsChange(ctx context.Context, sb []string, prosody []*ssml.Prosody) []float64 {
	res := make([]float64, len(sb))
	lastIsEmphasy := isEmphasyLast(prosody)
	isOneAccent, accentPos := isOneAccent(sb)
	if lastIsEmphasy && isOneAccent {
		// todo do separately
		log.Ctx(ctx).Debug().Int("pos", accentPos).Bool("emphasy", lastIsEmphasy).Msg("One accent with emphasy")
	}

	f := calculateSpeed(prosody)
	for i := 0; i < len(sb); i++ {
		res[i] = f
	}
	return res
}

func isOneAccent(sb []string) (bool, int) {
	res := -1
	for i, s := range sb {
		if strings.Contains(s, "\"") || strings.Contains(s, "^") {
			if res != -1 {
				return false, -1
			}
			res = i
		}
	}
	if res != -1 {
		return true, res
	}
	return false, -1
}

func isEmphasyLast(prosody []*ssml.Prosody) bool {
	if len(prosody) == 0 {
		return false
	}
	last := prosody[len(prosody)-1]
	return last.Emphasis == ssml.EmphasisTypeStrong || last.Emphasis == ssml.EmphasisTypeModerate || last.Emphasis == ssml.EmphasisTypeReduced
}

func calculateSpeed(prosody []*ssml.Prosody) float64 {
	total := 1.0
	for _, p := range prosody {
		nr := getRate(p)
		total *= nr
	}
	if total < 0.5 {
		total = 0.5
	} else if total > 2.0 {
		total = 2.0
	}
	return total
}

func getRate(p *ssml.Prosody) float64 {
	switch p.Emphasis {
	case ssml.EmphasisTypeReduced:
		return 1 / 1.3
	case ssml.EmphasisTypeNone:
		return 1.0
	case ssml.EmphasisTypeModerate:
		return 1.3
	case ssml.EmphasisTypeStrong:
		return 1.3 * 1.3
	default:
		if p.Rate <= 0 {
			return 1.0
		}
		return p.Rate
	}
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
