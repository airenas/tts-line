package processor

import (
	"context"
	"math"
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
	httpWrap      HTTPInvokerJSON
	url           string
	spaceSymbol   string
	endSymbol     string
	emphasisPause string

	hasVocoder bool
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
	res.emphasisPause = config.GetString("emphasisPause")
	if res.emphasisPause == "" {
		res.emphasisPause = "sp"
	}
	res.endSymbol = config.GetString("endSymbol")
	if res.endSymbol == "" {
		res.endSymbol = res.spaceSymbol
	}
	res.hasVocoder = config.GetBool("hasVocoder")
	log.Info().Str("AM pause", res.spaceSymbol).Str("AM emphasis pause", res.emphasisPause).
		Str("AM end symbol", res.endSymbol).Msg("AM")
	log.Info().Bool("has vocoder", res.hasVocoder).Msg("AM")
	return res, nil
}

func (p *amodel) Process(ctx context.Context, data *synthesizer.TTSDataPart) error {
	ctx, span := utils.StartSpan(ctx, "amodel.Process")
	defer span.End()

	if data.Cfg.Input.OutputFormat == api.AudioNone {
		if data.Cfg.Input.OutputTextFormat == api.TextTranscribed {
			inData, _, _ := p.mapAMInput(ctx, data)
			data.TranscribedText = inData.Text
		}
		return nil
	}

	inData, inIndexes, volChanges := p.mapAMInput(ctx, data)
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
	if err := mapAMOutputDurations(ctx, data, output.Durations, inIndexes, volChanges); err != nil {
		return err
	}

	if p.hasVocoder {
		data.Audio = output.Data
	} else {
		data.Spectogram = output.Data
	}
	return nil
}

func mapAMOutputDurations(ctx context.Context, data *synthesizer.TTSDataPart, durations []int, indRes []*synthesizer.SynthesizedPos, volChanges []float64) error {
	sums := make([]int, len(durations)+1)
	for i := 0; i < len(durations); i++ {
		sums[i+1] = sums[i] + durations[i]
	}
	for i, w := range data.Words {
		fromI := indRes[i].From
		toI := indRes[i].To
		if fromI < 0 || fromI >= len(sums) || toI < 0 || toI >= len(sums) {
			log.Ctx(ctx).Warn().Int("from", fromI).Int("to", toI).Int("len", len(sums)).Int("indResLen", len(indRes)).
				Msg("Invalid duration index")
			continue
		}
		if fromI >= len(volChanges) || toI > len(volChanges) {
			log.Ctx(ctx).Warn().Int("from", fromI).Int("to", toI).Int("len", len(volChanges)).
				Msg("Invalid volume change index")
			continue
		}
		w.SynthesizedPos = &synthesizer.SynthesizedPos{
			From:          sums[fromI],
			To:            sums[toI],
			Durations:     durations[fromI:toI],
			StartIndex:    fromI,
			VolumeChanges: volChanges[fromI:toI],
		}
	}
	return nil
}

type syllPos int

const (
	syllPosUnset syllPos = iota
	syllPosBeforeAccent
	syllPosAccent
	syllPosAfterAccent
)

type emphasisSummary struct {
	countTotal    int
	countAccented int
}

type syllInfo struct {
	textPart        *synthesizer.TTSTextPart
	syllPos         syllPos
	emphasisSummary *emphasisSummary
}

type amChar struct {
	char     string
	syllInfo *syllInfo
}

type amChars struct {
	items     []*amChar
	prepared  bool
	baseSpeed float64
}

func (si *syllInfo) getSyllPos() syllPos {
	if si == nil {
		return syllPosUnset
	}
	if si.emphasisSummary == nil || si.emphasisSummary.countAccented != 1 {
		return syllPosUnset
	}
	return si.syllPos
}

func (si *syllInfo) getProsodies() []*ssml.Prosody {
	if si == nil || si.textPart == nil {
		return nil
	}
	return si.textPart.Prosodies
}

func (a *amChars) pitch(ctx context.Context) [][]*syntmodel.PitchChange {
	a.prepare()

	res := make([][]*syntmodel.PitchChange, len(a.items))
	for i, c := range a.items {
		res[i] = makePitchChange(c.syllInfo.getProsodies(), c.syllInfo.getSyllPos())
	}
	return res
}

func (a *amChars) prepare() {
	if a.prepared {
		return
	}
	em := make(map[int][]*syllInfo)
	for _, c := range a.items {
		if c.syllInfo == nil {
			continue
		}
		if last, ok := lastEmphasy(c.syllInfo.getProsodies()); ok {
			id := last.ID
			sylls, ok := em[id]
			if !ok {
				sylls = make([]*syllInfo, 0, 1)
			}
			if len(sylls) == 0 || sylls[len(sylls)-1] != c.syllInfo {
				sylls = append(sylls, c.syllInfo)
			}
			em[id] = sylls
		}
	}
	for _, sylls := range em {
		emphasisSummary := &emphasisSummary{}
		acc := false
		for _, s := range sylls {
			s.emphasisSummary = emphasisSummary
			if s.syllPos == syllPosAccent {
				acc = true
				emphasisSummary.countAccented++
			} else if acc {
				s.syllPos = syllPosAfterAccent
			} else {
				s.syllPos = syllPosBeforeAccent
			}
			s.emphasisSummary.countTotal++
		}
	}
	a.prepared = true
}

func (a *amChars) durations(ctx context.Context) []float64 {
	a.prepare()

	res := make([]float64, len(a.items))
	for i, c := range a.items {
		res[i] = calculateSpeed(a.baseSpeed, c.syllInfo.getProsodies(), c.syllInfo.getSyllPos())
	}
	return res
}

func (a *amChars) volumes(ctx context.Context) []float64 {
	a.prepare()

	res := make([]float64, len(a.items))
	for i, c := range a.items {
		res[i] = calculateVolumeChange(c.syllInfo.getProsodies(), c.syllInfo.getSyllPos())
	}
	return res
}

func (a *amChars) text() string {
	res := strings.Builder{}
	for _, c := range a.items {
		if res.Len() > 0 {
			res.WriteString(" ")
		}
		res.WriteString(c.char)
	}
	return res.String()
}

func (a *amChars) endsWith(s string) bool {
	if len(a.items) == 0 {
		return false
	}
	parts := strings.Split(s, " ")
	if len(parts) == 0 || len(parts) > len(a.items) {
		return false
	}
	for i := 1; i <= len(parts); i++ {
		if a.items[len(a.items)-i].char != parts[len(parts)-i] {
			return false
		}
	}
	return true
}

func (a *amChars) add(s string, si *syllInfo) {
	parts := strings.Split(s, " ")
	for _, p := range parts {
		a.items = append(a.items, &amChar{
			char:     p,
			syllInfo: si,
		})
	}
}

func (a *amChars) replace(s, new string, si *syllInfo) {
	parts := strings.Split(s, " ")
	l := len(parts)
	if len(a.items) < l {
		return
	}
	for i := 0; i < l; i++ {
		a.items[len(a.items)-i-1] = nil
	}
	a.items = a.items[:len(a.items)-l]
	a.add(new, si)
}

func (p *amodel) mapAMInput(ctx context.Context, data *synthesizer.TTSDataPart) (*syntmodel.AMInput, []*synthesizer.SynthesizedPos, []float64) {
	res := &syntmodel.AMInput{}
	res.Voice = data.Cfg.Voice
	res.Priority = data.Cfg.Input.Priority
	if data.Cfg.JustAM {
		res.Text = data.Text
		return res, nil, nil
	}
	sb := amChars{baseSpeed: data.Cfg.Input.Speed}
	pause := p.spaceSymbol

	sb.add(pause, nil)
	lastSep := ""
	indRes := make([]*synthesizer.SynthesizedPos, len(data.Words))
	for i, w := range data.Words {
		indRes[i] = &synthesizer.SynthesizedPos{}
		indRes[i].From = len(sb.items)
		tgw := w.Tagged
		if tgw.Space {
			if w.IsLastInPart && w.TextPart != nil && w.TextPart.PauseAfter > 0 {
				sb.add("sil", &syllInfo{textPart: w.TextPart})
			}
		} else if tgw.Separator != "" {
			sep := getSep(tgw.Separator, data.Words, i)
			if sep != "" {
				sb.add(sep, &syllInfo{textPart: w.TextPart})
				lastSep = sep
			}
			if addPause(sep, data.Words, i) {
				sb.add(pause, &syllInfo{textPart: w.TextPart})
			}
		} else if tgw.SentenceEnd {
			if lastSep == "" {
				lastSep = "."
				sb.add(lastSep, &syllInfo{textPart: w.TextPart})
			}
			if !sb.endsWith(pause) {
				sb.add(pause, &syllInfo{textPart: w.TextPart})
			}
		} else {
			phns := strings.Split(w.Transcription, " ")
			si := &syllInfo{textPart: w.TextPart}
			for _, p := range phns {
				if p == "-" {
					si = &syllInfo{textPart: w.TextPart}
				}
				pt := changePhn(p)
				if pt != "" {
					if accented(pt) {
						si.syllPos = syllPosAccent
					}
					sb.add(pt, si)
					lastSep = ""
				}
			}
			// add pause if it is the last word of an emphasis
			if w.LastEmphasisWord {
				sb.add(p.emphasisPause, si)
			}
		}
		indRes[i].To = len(sb.items)
	}

	l := len(sb.items)
	if l > 0 {
		if !sb.endsWith(p.endSymbol) {
			if sb.endsWith(pause) {
				sb.replace(pause, p.endSymbol, nil)
			} else if sb.endsWith(p.emphasisPause) {
				sb.replace(p.emphasisPause, p.endSymbol, nil)
			} else {
				sb.add(p.endSymbol, nil)
			}
		}
	}
	if len(indRes) > 0 {
		lEnd := len(strings.Split(p.endSymbol, " "))
		if lEnd > 1 {
			indRes[len(indRes)-1].To += (lEnd - 1)
		}
	}
	res.Text = sb.text()
	res.DurationsChange = sb.durations(ctx)
	res.PitchChange = sb.pitch(ctx)
	return res, indRes, sb.volumes(ctx)
}

func accented(pt string) bool {
	return strings.Contains(pt, "\"") || strings.Contains(pt, "^")
}

func makePitchChange(prosody []*ssml.Prosody, sp syllPos) []*syntmodel.PitchChange {
	res := make([]*syntmodel.PitchChange, 0, len(prosody))
	for i, p := range prosody {
		last := i == len(prosody)-1
		switch p.Emphasis {
		case ssml.EmphasisTypeModerate:
			pc := &syntmodel.PitchChange{
				Value: getPitchMultiplierForEphasis(sp, last),
				Type:  2,
			}
			res = append(res, pc)
		case ssml.EmphasisTypeStrong:
			pc := &syntmodel.PitchChange{
				Value: 1.1 * getPitchMultiplierForEphasis(sp, last),
				Type:  2,
			}
			res = append(res, pc)
		case ssml.EmphasisTypeReduced:
			pc := &syntmodel.PitchChange{
				Value: 1 / getPitchMultiplierForEphasis(sp, last),
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

func getPitchMultiplierForEphasis(sp syllPos, last bool) float64 {
	if sp == syllPosAccent && last {
		return 1.2
	}
	return 1.1
}

func lastEmphasy(prosody []*ssml.Prosody) (*ssml.Prosody, bool) {
	if len(prosody) == 0 {
		return nil, false
	}
	last := prosody[len(prosody)-1]
	if last.Emphasis == ssml.EmphasisTypeUnset || last.Emphasis == ssml.EmphasisTypeNone {
		return nil, false
	}
	return last, true
}

func calculateSpeed(baseSpeed float64, prosody []*ssml.Prosody, sp syllPos) float64 {
	total := baseSpeed
	if total <= 0 {
		total = 1.0
	}
	for i, p := range prosody {
		last := i == len(prosody)-1
		if !last {
			sp = syllPosUnset
		}
		nr := getRate(p, sp)
		total *= nr
	}
	if total < 0.5 {
		total = 0.5
	} else if total > 2.0 {
		total = 2.0
	}
	return total
}

func calculateVolumeChange(prosody []*ssml.Prosody, sp syllPos) float64 {
	res := 0.0
	for i, p := range prosody {
		if utils.Float64Equals(p.Volume, ssml.MinVolumeChange) {
			return ssml.MinVolumeChange
		}
		last := i == len(prosody)-1
		if !last {
			sp = syllPosUnset
		}
		res += getVolumeChange(p, sp)
	}
	return res
}

func getVolumeChange(p *ssml.Prosody, sp syllPos) float64 {
	step := getEmphasisVolChangeStep(sp) // dB
	switch p.Emphasis {
	case ssml.EmphasisTypeReduced:
		return -step
	case ssml.EmphasisTypeNone:
		return 0.0
	case ssml.EmphasisTypeModerate:
		return step
	case ssml.EmphasisTypeStrong:
		return 2 * step
	default:
		return p.Volume
	}
}

func getEmphasisVolChangeStep(sp syllPos) float64 {
	if sp == syllPosUnset {
		return 2.0
	}
	if sp == syllPosAccent {
		return 3.0
	}
	return 0
}

func calcVolumeRate(changeInDB float64) float64 {
	if utils.Float64Equals(changeInDB, ssml.MinVolumeChange) {
		return 0
	}
	return math.Pow(10, changeInDB/20)
}

func getRate(p *ssml.Prosody, sp syllPos) float64 {
	er := getEmphasisRate(sp)
	switch p.Emphasis {
	// case ssml.EmphasisTypeReduced:  // leave reate as it is
	// 	// reduce speed
	// 	return 1 / er
	case ssml.EmphasisTypeNone:
		return 1.0
	case ssml.EmphasisTypeModerate:
		return er
	case ssml.EmphasisTypeStrong:
		return 1.1 * er
	default:
		if p.Rate <= 0 {
			return 1.0
		}
		return p.Rate
	}
}

func getEmphasisRate(sp syllPos) float64 {
	switch sp {
	case syllPosBeforeAccent:
		return 1.3
	case syllPosAfterAccent:
		return 1.2
	case syllPosAccent:
		return 1.4
	}
	return 1.2
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
	if words[pos].Tagged.Mi == dashMI {
		s = "-"
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
	if words[pos].Tagged.Mi == dashMI {
		s = "-"
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
