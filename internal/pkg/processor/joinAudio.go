package processor

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// AudioLoader provides wav data by key
type AudioLoader interface {
	TakeWav(string) ([]byte, error)
}

type joinAudio struct {
	suffixProvider AudioLoader
}

// NewJoinAudio joins results into one audio
func NewJoinAudio(suffixProvider AudioLoader) synthesizer.Processor {
	return &joinAudio{suffixProvider: suffixProvider}
}

func (p *joinAudio) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "joinAudio.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var suffix []byte
	var err error
	if data.AudioSuffix != "" {
		if suffix, err = p.suffixProvider.TakeWav(data.AudioSuffix); err != nil {
			return errors.Wrapf(err, "can't take suffix %s", data.AudioSuffix)
		}
	}
	for _, p := range data.Parts {
		p.TranscribedSymbols = strings.Split(p.TranscribedText, " ")
	}

	data.Audio, data.AudioLenSeconds, data.SampleRate, err = join(ctx, data.Parts, suffix, data.Input.MaxEdgeSilenceMillis)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData(ctx, "Output", fmt.Sprintf("audio len %d", len(data.Audio)), nil)
	return nil
}

type wavWriter struct {
	header         []byte
	size           uint32
	buf            bytes.Buffer
	sampleRateV    uint32
	bitsPerSampleV uint16
}

func (wr *wavWriter) init(wavData []byte) {
	if wr.header != nil {
		return
	}
	wr.header = wav.TakeHeader(wavData)
}

func (wr *wavWriter) sampleRate() uint32 {
	if wr.sampleRateV != 0 {
		return wr.sampleRateV
	}
	if wr.header == nil {
		return 0
	}
	wr.sampleRateV = wav.GetSampleRate(wr.header)
	return wr.sampleRateV
}

func (wr *wavWriter) bitsPerSample() uint16 {
	if wr.bitsPerSampleV != 0 {
		return wr.bitsPerSampleV
	}
	if wr.header == nil {
		return 0
	}
	wr.bitsPerSampleV = wav.GetBitsPerSample(wr.header)
	return wr.bitsPerSampleV
}

func join(ctx context.Context, parts []*synthesizer.TTSDataPart, suffix []byte, maxEdgeSilenceMilis int64) ([]byte, float64, uint32 /*sampleRate*/, error) {
	res := &wavWriter{}
	nextStartSil := 0
	for i, part := range parts {
		decoded := part.Audio
		res.init(decoded)
		startSil, endSil := nextStartSil, 0
		if i == 0 {
			startSil = getStartSilSize(part.TranscribedSymbols, part.Durations)
			startDurHops := part.DefaultSilence / 2
			if maxEdgeSilenceMilis > -1 {
				startDurHops = min(startDurHops, toHops(maxEdgeSilenceMilis, part.Step, res.sampleRate()))
			}
			startSil, _, _ = calcPauseWithEnds(startSil, 0, startDurHops)
		}
		if i == len(parts)-1 {
			endSil = getEndSilSize(ctx, part.TranscribedSymbols, part.Durations)
			if suffix != nil {
				_, endSil, _ = calcPauseWithEnds(0, endSil, part.DefaultSilence)
			} else {
				endDurHops := part.DefaultSilence / 2
				if maxEdgeSilenceMilis > -1 {
					endDurHops = min(endDurHops, toHops(maxEdgeSilenceMilis, part.Step, res.sampleRate()))
				}
				_, endSil, _ = calcPauseWithEnds(0, endSil, endDurHops)
			}
		} else if i < len(parts)-1 {
			endSil = getEndSilSize(ctx, part.TranscribedSymbols, part.Durations)
			nextStartSil = getStartSilSize(parts[i+1].TranscribedSymbols, parts[i+1].Durations)
			endSil, nextStartSil, _ = calcPauseWithEnds(endSil, nextStartSil, part.DefaultSilence)
		}
		startSkip, endSkip := startSil*part.Step, endSil*part.Step
		lenBefore := res.buf.Len()
		if err := appendWav(res, decoded, startSkip, endSkip); err != nil {
			return nil, 0, 0, err
		}
		part.AudioDurations = &synthesizer.AudioDurations{
			Shift:    -startSil,
			Duration: calculateDurations(res.buf.Len()-lenBefore, res.sampleRate()*uint32(res.bitsPerSample())/uint32(8)),
		}
	}
	if res.size == 0 {
		return nil, 0, 0, errors.New("no wav data")
	}
	if suffix != nil {
		if err := appendWav(res, suffix, 0, 0); err != nil {
			return nil, 0, 0, errors.Wrapf(err, "can't append suffix")
		}
	}
	var bufRes bytes.Buffer
	_, _ = bufRes.Write(res.header)

	_, _ = bufRes.Write([]byte("data"))
	_, _ = bufRes.Write(wav.SizeBytes(res.size))
	_, _ = bufRes.Write(res.buf.Bytes())
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return nil, 0, 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.Bytes(), float64(res.size) / float64(bitsRate), res.sampleRate(), nil
}

func calculateDurations(aLen int, samplesPerSec uint32) time.Duration {
	if samplesPerSec == 0 {
		return 0
	}
	return time.Duration(float64(aLen)*1000/float64(samplesPerSec)) * time.Millisecond
}

func getStartSilSize(phones []string, durations []int) int {
	l := len(phones)
	res := 0
	for i := 0; i < l-3 && i < len(durations); i++ {
		if isSil(phones[i]) {
			res = res + durations[i]
		} else {
			return res
		}
	}
	return res
}

func isSil(ph string) bool {
	return ph == "sil" || ph == "sp" || (len(ph) == 1 && unicode.IsPunct([]rune(ph)[0]))
}

func getEndSilSize(ctx context.Context, phones []string, durations []int) int {
	l := len(phones)
	if len(durations) != l+1 {
		log.Ctx(ctx).Warn().Msg("Duration size don't match phone list")
		return 0
	}
	res := durations[l]
	for i := l - 1; i > 1; i-- {
		if isSil(phones[i]) {
			res = res + durations[i]
		} else {
			return res
		}
	}
	return res
}

func appendWav(res *wavWriter, wavData []byte, startSkip, endSkip int) error {
	if !wav.IsValid(wavData) {
		return errors.New("no valid audio wave data")
	}
	header := wav.TakeHeader(wavData)
	if res.header != nil {
		if res.sampleRate() != wav.GetSampleRate(header) {
			return errors.Errorf("differs sample rate %d vs %d", res.sampleRate(), wav.GetSampleRate(header))
		}
		if res.bitsPerSample() != wav.GetBitsPerSample(header) {
			return errors.Errorf("differs bits per sample %d vs %d", res.bitsPerSample(), wav.GetBitsPerSample(header))
		}
	}
	bData := wav.TakeData(wavData)
	bps := res.bitsPerSample()
	fix := int(bps / 8)
	es := len(bData) - (endSkip * fix)
	if es < (startSkip * fix) {
		return errors.Errorf("audio start pos > end: %d vs %d", startSkip, es)
	}

	res.size += uint32(es - (startSkip * fix))

	from, to := startSkip*fix, es
	_, err := res.buf.Write(bData[from:to])
	return err
}

// Info return info about processor
func (p *joinAudio) Info() string {
	return fmt.Sprintf("joinAudio(%s)", utils.RetrieveInfo(p.suffixProvider))
}

type joinSSMLAudio struct {
	suffixProvider AudioLoader
}

// NewJoinSSMLAudio joins results into one audio from many ssml parts
func NewJoinSSMLAudio(suffixProvider AudioLoader) synthesizer.Processor {
	return &joinSSMLAudio{suffixProvider: suffixProvider}
}

func (p *joinSSMLAudio) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "joinSSMLAudio.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var suffix []byte
	var err error
	if data.AudioSuffix != "" {
		if suffix, err = p.suffixProvider.TakeWav(data.AudioSuffix); err != nil {
			return errors.Wrapf(err, "can't take suffix %s", data.AudioSuffix)
		}
	}
	for _, dp := range data.SSMLParts {
		if dp.Cfg.Type == synthesizer.SSMLText {
			for _, p := range dp.Parts {
				p.TranscribedSymbols = strings.Split(p.TranscribedText, " ")
			}
		}
	}
	data.Audio, data.AudioLenSeconds, data.SampleRate, err = joinSSML(ctx, data, suffix, data.Input.MaxEdgeSilenceMillis)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	for _, dp := range data.SSMLParts {
		dp.SampleRate = data.SampleRate
	}
	utils.LogData(ctx, "Output", fmt.Sprintf("audio len %d", len(data.Audio)), nil)
	return nil
}

type nextWriteData struct {
	part        *synthesizer.TTSDataPart
	data        []byte
	startSkip   int
	pause       time.Duration // pause that should be written after data
	isPause     bool
	durationAdd time.Duration // time to add to the part
	volChanges  []volChange
}

type volChange struct {
	from   int
	to     int
	change float64
}

func joinSSML(ctx context.Context, data *synthesizer.TTSData, suffix []byte, maxEdgeSilenceMillis int64) ([]byte, float64 /*len*/, uint32 /*sampleRate*/, error) {
	res := &wavWriter{}
	wd := &nextWriteData{}
	wd.pause = time.Duration(0)
	first := true
	writeF := func(part *synthesizer.TTSDataPart, last bool) error {
		step, defaultSil, pause := 0, 0, time.Duration(0)
		endSil, startSil := 0, 0
		var decoded []byte
		var err error
		if part != nil {
			decoded = part.Audio
			res.init(decoded)
			if err != nil {
				return err
			}
			if !wav.IsValid(decoded) {
				return errors.New("no valid audio wave data")
			}
			if res.header == nil {
				res.header = wav.TakeHeader(decoded)
			}
			startSil = getStartSilSize(part.TranscribedSymbols, part.Durations)
			step = part.Step
			defaultSil = part.DefaultSilence
			if first && maxEdgeSilenceMillis > -1 {
				defaultSil = min(defaultSil, toHops(maxEdgeSilenceMillis, part.Step, res.sampleRate()))
			}
			first = false
		}
		if wd.part != nil {
			endSil = getEndSilSize(ctx, wd.part.TranscribedSymbols, wd.part.Durations)
			step = wd.part.Step
			defaultSil = wd.part.DefaultSilence
			if last && maxEdgeSilenceMillis > -1 {
				defaultSil = min(defaultSil, toHops(maxEdgeSilenceMillis, wd.part.Step, res.sampleRate()))
			}
		}
		pause = 0
		if wd.isPause {
			startSil, endSil, pause = fixPause(startSil, endSil, wd.pause, step, res.sampleRate())
		} else {
			startSil, endSil, _ = calcPauseWithEnds(startSil, endSil, defaultSil)
		}

		if step == 0 {
			return fmt.Errorf("no step")
		}

		startSkip, endSkip := startSil*step, endSil*step
		if wd.part != nil {
			lenBefore := res.buf.Len()
			if err := appendWav(res, wd.data, wd.startSkip, endSkip); err != nil {
				return err
			}
			wd.part.AudioDurations = &synthesizer.AudioDurations{
				Shift:    -wd.startSkip / step,
				Duration: calculateDurations(res.buf.Len()-lenBefore, res.sampleRate()*uint32(res.bitsPerSample())/uint32(8)),
			}
			wd.part.AudioDurations.Shift += utils.ToTTSSteps(wd.durationAdd, wav.GetSampleRate(res.header), step)
			wd.part.AudioDurations.Duration += wd.durationAdd
			wd.durationAdd = 0
			volumeChange := calcVolumeChange(wd.part.Cfg.Prosodies)
			if !utils.Float64Equals(volumeChange, 0) {
				log.Ctx(ctx).Warn().Float64("change", volumeChange).Int("from", lenBefore).Int("to", res.buf.Len()).Msg("volume")
				wd.volChanges = append(wd.volChanges, volChange{
					from:   lenBefore,
					to:     res.buf.Len(),
					change: calcVolumeRate(volumeChange),
				})
			}
		}
		if pause > 0 {
			if err := appendPause(ctx, res, pause); err != nil {
				return err
			}
			wd.durationAdd = pause
			wd.pause = 0
		}
		wd.isPause = false
		wd.part = part
		wd.startSkip = startSkip
		wd.data = decoded
		return nil
	}

	for _, dp := range data.SSMLParts {
		switch dp.Cfg.Type {
		case synthesizer.SSMLPause:
			wd.pause = wd.pause + dp.Cfg.PauseDuration
			wd.isPause = true
		case synthesizer.SSMLText:
			for _, part := range dp.Parts {
				err := writeF(part, false /*last*/)
				if err != nil {
					return nil, 0, 0, err
				}
			}
		}
	}
	if err := writeF(nil, suffix == nil /*last*/); err != nil {
		return nil, 0, 0, err
	}
	if res.size == 0 {
		return nil, 0, 0, errors.New("no audio")
	}
	if suffix != nil {
		if err := appendWav(res, suffix, wd.startSkip, 0); err != nil {
			return nil, 0, 0, errors.Wrapf(err, "can't append suffix")
		}
	}

	resBytes, err := changeVolume(res.buf.Bytes(), wd.volChanges, int(res.bitsPerSample()/8))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("change volume: %w", err)
	}
	var bufRes bytes.Buffer
	_, _ = bufRes.Write(res.header)
	_, _ = bufRes.Write([]byte("data"))
	_, _ = bufRes.Write(wav.SizeBytes(res.size))
	_, _ = bufRes.Write(resBytes)
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return nil, 0, 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.Bytes(), float64(res.size) / float64(bitsRate), res.sampleRate(), nil
}

func changeVolume(b []byte, volChange []volChange, bytesPerSaample int) ([]byte, error) {
	if len(volChange) == 0 {
		return b, nil
	}
	for _, vc := range volChange {
		for i := vc.from; i < vc.to; i += bytesPerSaample {
			switch bytesPerSaample {
			case 2:
				sample := int16(b[i]) | int16(b[i+1])<<8
				newSample := toInt16(float64(sample) * vc.change)
				b[i] = byte(newSample & 0xFF)
				b[i+1] = byte((newSample >> 8) & 0xFF)
			case 4:
				return nil, fmt.Errorf("not implemented for 4 bytes per sample")
			default:
				return nil, fmt.Errorf("unsupported bytes per sample %d", bytesPerSaample)
			}
		}
	}
	return b, nil
}

func toInt16(f float64) int16 {
	if f > 32767 {
		return 32767
	}
	if f < -32768 {
		return -32768
	}
	return int16(math.Round(f))
}

func calcVolumeChange(prosody []*ssml.Prosody) float64 {
	res := 0.0
	for _, p := range prosody {
		if utils.Float64Equals(p.Volume, ssml.MinVolumeChange) {
			return ssml.MinVolumeChange
		}
		res += getVolumeChange(p)
	}
	return res
}

func getVolumeChange(p *ssml.Prosody) float64 {
	step := 2.0 // dB
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

func calcVolumeRate(changeInDB float64) float64 {
	if utils.Float64Equals(changeInDB, ssml.MinVolumeChange) {
		return 0
	}
	return math.Pow(10, changeInDB/20)
}

func calcPauseWithEnds(s1, s2, pause int) (int, int, int) {
	if s1+s2 <= pause {
		return 0, 0, pause - s1 - s2
	}
	if s1 < s2 {
		r := min(s1, pause/2)
		return s1 - r, s2 - max(pause-r, 0), 0
	}
	r := min(s2, pause/2)
	return s1 - max(pause-r, 0), s2 - r, 0
}

func fixPause(s1, s2 int, pause time.Duration, step int, sampleRate uint32) (int, int, time.Duration) {
	if step == 0 {
		return s1, s2, pause
	}
	millisInHops := 1000 * float64(step) / float64(sampleRate)
	pauseHops := toHops(pause.Milliseconds(), step, sampleRate)
	r1, r2, rp := calcPauseWithEnds(s1, s2, pauseHops)
	return r1, r2, time.Millisecond * time.Duration(int(float64(rp)*millisInHops))
}

func toHops(millis int64, step int, sampleRate uint32) int {
	if step == 0 {
		return 0
	}
	res := int(math.Round(float64(millis) * float64(sampleRate) / 1000 / float64(step)))
	return res
}

func appendPause(ctx context.Context, res *wavWriter, pause time.Duration) error {
	if res.header == nil {
		return errors.New("no wav data before pause")
	}
	c, err := writePause(ctx, &res.buf, res.sampleRate(), res.bitsPerSample(), pause)
	if err != nil {
		return err
	}
	res.size += c
	return nil
}

func writePause(ctx context.Context, buf *bytes.Buffer, sampleRate uint32, bitsPerSample uint16, pause time.Duration) (uint32, error) {
	if pause > time.Second*10 {
		log.Ctx(ctx).Warn().Msgf("Too long pause %v", pause)
		pause = time.Second * 10
	}
	if pause < 0 {
		pause = 0
	}
	c := uint32(pause.Milliseconds()*int64(sampleRate)/1000) * uint32(bitsPerSample/8)
	for i := uint32(0); i < c; i++ {
		if err := buf.WriteByte(0); err != nil {
			return 0, err
		}
	}
	return c, nil
}

// Info return info about processor
func (p *joinSSMLAudio) Info() string {
	return fmt.Sprintf("joinSSMLAudio(%s)", utils.RetrieveInfo(p.suffixProvider))
}
