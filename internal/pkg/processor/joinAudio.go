package processor

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/tts-line/internal/pkg/audio"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
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

	data.Audio, err = join(ctx, data.Parts, suffix, data.Input.MaxEdgeSilenceMillis)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData(ctx, "Output", fmt.Sprintf("audio len %d", len(data.Audio.Data)), nil)
	return nil
}

type wavWriter struct {
	header []byte
	// size           uint32
	buf            bytes.Buffer
	sampleRateV    uint32
	bitsPerSampleV uint16

	maxEdgeSilenceMillis int64
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

func (wr *wavWriter) bytesPerSample() uint16 {
	return wr.bitsPerSample() / 8
}

func join(ctx context.Context, parts []*synthesizer.TTSDataPart, suffix []byte, maxEdgeSilenceMilis int64) (*synthesizer.AudioData, error) {
	ctx, span := utils.StartSpan(ctx, "joinAudio.join")
	defer span.End()

	res := &wavWriter{maxEdgeSilenceMillis: maxEdgeSilenceMilis}

	var volChanges []*audio.VolChange
	//prealocate data

	// add words
	wwd, wwdNext := &wordWriteData{}, &wordWriteData{}
	for _, part := range parts {
		ar, err := initAudioReader(ctx, part)
		if err != nil {
			return nil, err
		}
		if res.header == nil {
			res.header = ar.audio.header
			res.bitsPerSampleV = ar.audio.bitsPerSample
			res.sampleRateV = ar.audio.sampleRate
		}
		lenBefore := res.buf.Len()
		if len(part.Words) == 0 {
			// just append all
			bData := ar.audio.data
			_, err := res.buf.Write(bData)
			if err != nil {
				return nil, err
			}
			continue
		}
		for _, w := range part.Words {
			wwdNext.audioReader = ar
			wwdNext.word = w
			err := writeWordAudio(ctx, res, wwd, wwdNext)
			if err != nil {
				return nil, err
			}
			wwd, wwdNext = wwdNext, &wordWriteData{}
		}
		vc := makeVolumeChanges(ctx, part, lenBefore, res.buf.Len(), res.bytesPerSample())
		if len(vc) > 0 {
			volChanges = append(volChanges, vc...)
		}
	}
	err := writeWordAudio(ctx, res, wwd, wwdNext)
	if err != nil {
		return nil, err
	}

	if res.buf.Len() == 0 {
		return nil, errors.New("no audio")
	}

	if suffix != nil {
		if err := appendWav(ctx, res, suffix); err != nil {
			return nil, errors.Wrapf(err, "can't append suffix")
		}
	}

	resBytes, err := audio.ChangeVolume(ctx, res.buf.Bytes(), volChanges, int(res.bytesPerSample()))
	if err != nil {
		return nil, fmt.Errorf("change volume: %w", err)
	}

	var bufRes bytes.Buffer
	_, _ = bufRes.Write(res.header)

	_, _ = bufRes.Write([]byte("data"))
	_, _ = bufRes.Write(wav.SizeBytes(uint32(res.buf.Len())))
	_, _ = bufRes.Write(resBytes)
	return &synthesizer.AudioData{
		Data:          bufRes.Bytes(),
		SampleRate:    res.sampleRate(),
		BitsPerSample: res.bitsPerSample(),
	}, nil

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

func getSilSize(phones []string, durations []int, at, move int) int {
	l := len(phones)
	res := 0
	for i := at; i < l && i >= 0 && utils.Abs(i-at) < 3; i += move {
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

func getEndSilSize(ctx context.Context, phones []string, durations []int) (int /*count*/, int /*at*/) {
	l := len(phones)
	if len(durations) != l+1 {
		log.Ctx(ctx).Warn().Msg("Duration size don't match phone list")
		return 0, 0
	}
	res := durations[l]
	sum := 0
	for _, d := range durations[:l] {
		sum += d
	}
	for i := l - 1; i > 1; i-- {
		if isSil(phones[i]) {
			res = res + durations[i]
			sum -= durations[i]
		} else {
			return res, sum
		}
	}
	return res, sum
}

func appendWav(_ctx context.Context, res *wavWriter, wavData []byte) error {
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
	_, err := res.buf.Write(bData)
	if err != nil {
		return err
	}
	return nil
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
	data.Audio, err = joinSSML(ctx, data, suffix, data.Input.MaxEdgeSilenceMillis)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	for _, dp := range data.SSMLParts {
		dp.Audio = data.Audio
	}
	utils.LogData(ctx, "Output", fmt.Sprintf("audio len %d", len(data.Audio.Data)), nil)
	return nil
}

type parsedWAV struct {
	header        []byte
	data          []byte
	sampleRate    uint32
	bitsPerSample uint16
}

type audioReader struct {
	part        *synthesizer.TTSDataPart
	audio       *parsedWAV
	wrote       int // bytes
	wrotePos    int // in samples
	step        int // in samples
	startSil    int // in samples
	endSil      int // in samples
	endSilStart int // in samples
}

func (a *audioReader) wordShiftInAudio(from int) int {
	shift := from - a.wrotePos
	return shift * a.step * int(a.audio.bitsPerSample) / 8
}

type wordWriteData struct {
	word        *synthesizer.ProcessedWord
	audioReader *audioReader
	silence     time.Duration
	cutSteps    int // in tts steps
}

func writeWordAudio(ctx context.Context, res *wavWriter, wwd *wordWriteData, wwdNext *wordWriteData) error {
	if wwd.word == nil && wwdNext.word == nil { // both nil ???
		return nil
	}

	if wwd.word == nil { // first
		silTTSSteps := wwdNext.audioReader.startSil
		silAtStart := utils.ToDuration(silTTSSteps, wwdNext.audioReader.audio.sampleRate, wwdNext.audioReader.step)
		if wwd.silence > silAtStart {
			if err := appendPause(ctx, res, wwd.silence-silAtStart); err != nil {
				return err
			}
			wwd.silence = 0
			return nil
		}
		defaultSil := wwdNext.audioReader.part.DefaultSilence / 2
		if res.maxEdgeSilenceMillis > -1 {
			defaultSil = toHops(res.maxEdgeSilenceMillis, wwdNext.audioReader.step, res.sampleRate())
		}
		if silTTSSteps > defaultSil {
			skip := silTTSSteps - defaultSil
			wwdNext.audioReader.wrotePos = skip
			wwdNext.audioReader.wrote = skip * wwdNext.audioReader.step * int(wwdNext.audioReader.audio.bitsPerSample) / 8
		}
		return nil
	}
	if wwdNext.word == nil { // last
		silTTSSteps := wwd.audioReader.endSil
		silAtEnd := utils.ToDuration(silTTSSteps, wwd.audioReader.audio.sampleRate, wwd.audioReader.step)
		if wwd.silence > silAtEnd {
			err := appendAudioBytes(ctx, res, wwd.audioReader, wwd.audioReader.endSilStart+wwd.audioReader.endSil)
			if err != nil {
				return err
			}
			if err := appendPause(ctx, res, wwd.silence-silAtEnd); err != nil {
				return err
			}
			wwd.silence = 0
			return nil
		}
		defaultSil := wwd.audioReader.part.DefaultSilence / 2
		if res.maxEdgeSilenceMillis > -1 {
			defaultSil = toHops(res.maxEdgeSilenceMillis, wwd.audioReader.step, res.sampleRate())
		}
		write := min(defaultSil, silTTSSteps)
		err := appendAudioBytes(ctx, res, wwd.audioReader, wwd.audioReader.endSilStart+write)
		if err != nil {
			return err
		}
		return nil
	}
	if wwd.audioReader.wrote > 0 && wwdNext.audioReader.wrote == 0 { // on parts boundary
		silTTSSteps := wwdNext.audioReader.startSil + wwd.audioReader.endSil
		silDuration := utils.ToDuration(silTTSSteps, wwdNext.audioReader.audio.sampleRate, wwdNext.audioReader.step)
		if wwd.silence > silDuration {
			err := appendAudioBytes(ctx, res, wwd.audioReader, wwd.audioReader.endSilStart+wwd.audioReader.endSil)
			if err != nil {
				return err
			}
			if err := appendPause(ctx, res, wwd.silence-silDuration); err != nil {
				return err
			}
			wwd.silence = 0
			return nil
		}
		defaultSil := wwdNext.audioReader.part.DefaultSilence
		if wwd.silence > 0 {
			defaultSil = toHops(wwd.silence.Milliseconds(), wwdNext.audioReader.step, wwdNext.audioReader.audio.sampleRate)
		}
		if silTTSSteps > defaultSil {
			endSkipSil, nextStartSil, _ := calcPauseWithEnds(wwd.audioReader.endSil, wwdNext.audioReader.startSil, defaultSil)
			err := appendAudioBytes(ctx, res, wwd.audioReader, wwd.audioReader.endSilStart+wwd.audioReader.endSil-endSkipSil)
			if err != nil {
				return err
			}
			wwdNext.audioReader.wrotePos = nextStartSil
			wwdNext.audioReader.wrote = nextStartSil * wwdNext.audioReader.step * int(wwdNext.audioReader.audio.bitsPerSample) / 8
		}
		return nil
	}

	sil := wwd.silence
	if sil > 0 {
		if err := appendPauseWithSearchBest(ctx, res, sil, wwd.audioReader, wwd.word.SynthesizedPos.From); err != nil {
			return err
		}
		wwd.silence = 0
	}
	if wwd.cutSteps > 0 {
		err := cutAudioBytes(ctx, res, wwd.audioReader, wwd.word.SynthesizedPos.From, wwd.cutSteps)
		if err != nil {
			return err
		}
		wwd.cutSteps = 0
	}

	wwd.word.AudioPos = &synthesizer.AudioPos{
		From: res.buf.Len() + wwd.audioReader.wordShiftInAudio(wwd.word.SynthesizedPos.From),
		To:   res.buf.Len() + wwd.audioReader.wordShiftInAudio(wwd.word.SynthesizedPos.To),
	}

	if wwd.word.Tagged.IsWord() {
		err := appendAudioBytes(ctx, res, wwd.audioReader, wwd.word.SynthesizedPos.To)
		if err != nil {
			return err
		}
	}
	if wwd.word.IsLastInPart && wwd.word.TextPart.PauseAfter > 0 {
		silTTSSteps := getSilSize(wwd.audioReader.part.TranscribedSymbols, wwd.audioReader.part.Durations, wwd.word.SynthesizedPos.StartIndex, -1)
		silDuration := utils.ToDuration(silTTSSteps, wwd.audioReader.audio.sampleRate, wwd.audioReader.step)
		if wwd.word.TextPart.PauseAfter >= silDuration {
			wwdNext.silence = wwd.word.TextPart.PauseAfter - silDuration
		} else {
			skipTTSSteps := toHops((silDuration - wwd.word.TextPart.PauseAfter).Milliseconds(), wwd.audioReader.step, wwd.audioReader.audio.sampleRate)
			wwdNext.cutSteps = skipTTSSteps
		}
	}

	return nil
}

func appendPauseWithSearchBest(ctx context.Context, res *wavWriter, sil time.Duration, audioReader *audioReader, to int) error {
	// find best pos to insert
	from := audioReader.wrotePos
	if to < from {
		return appendPause(ctx, res, sil)
	}

	a := audioReader.audio
	fromBytes := from * audioReader.step * int(a.bitsPerSample) / 8
	toBytes := to * audioReader.step * int(a.bitsPerSample) / 8

	window := (toBytes - fromBytes)
	center := fromBytes + window/2
	if window > 1100 {
		window = 1100
	}
	pos := audio.SearchBestSilAround(ctx, a.data, a.bitsPerSample/8, center, window)
	stepsToMove := (pos - fromBytes) / (audioReader.step * int(a.bitsPerSample) / 8)

	// move to best pos
	err := appendAudioBytes(ctx, res, audioReader, audioReader.wrotePos+stepsToMove)
	if err != nil {
		return err
	}
	return appendPause(ctx, res, sil)
}

func cutAudioBytes(ctx context.Context, res *wavWriter, audioReader *audioReader, to, steps int) error {
	// find best pos to cut
	from := audioReader.wrotePos
	if to < from {
		return nil
	}

	if from+steps >= to { // trim everything
		audioReader.wrote += (to - from) * audioReader.step * int(audioReader.audio.bitsPerSample) / 8
		audioReader.wrotePos += (to - from)
		return nil
	}

	a := audioReader.audio
	fromBytes := from * audioReader.step * int(a.bitsPerSample) / 8
	toBytes := (to - steps) * audioReader.step * int(a.bitsPerSample) / 8

	if fromBytes >= toBytes {
		return nil
	}

	window := (toBytes - fromBytes)
	center := fromBytes + window/2
	if window > 1100 {
		window = 1100
	}
	pos := audio.SearchBestSilAround(ctx, a.data, a.bitsPerSample/8, center, window)
	stepsToMove := (pos - fromBytes) / (audioReader.step * int(a.bitsPerSample) / 8)

	// move to the best pos
	err := appendAudioBytes(ctx, res, audioReader, audioReader.wrotePos+stepsToMove)
	if err != nil {
		return err
	}
	// now skip
	audioReader.wrote += steps * audioReader.step * int(audioReader.audio.bitsPerSample) / 8
	audioReader.wrotePos += steps
	return nil
}

func joinSSML(ctx context.Context, data *synthesizer.TTSData, suffix []byte, maxEdgeSilenceMillis int64) (*synthesizer.AudioData, error) {
	ctx, span := utils.StartSpan(ctx, "joinSSML")
	defer span.End()

	res := &wavWriter{maxEdgeSilenceMillis: maxEdgeSilenceMillis}

	var volChanges []*audio.VolChange
	//prealocate data

	// add words
	wwd, wwdNext := &wordWriteData{}, &wordWriteData{}
	for _, dp := range data.SSMLParts {
		switch dp.Cfg.Type {
		case synthesizer.SSMLPause:
			wwd.silence = wwd.silence + dp.Cfg.PauseDuration
		case synthesizer.SSMLText:
			for _, part := range dp.Parts {
				ar, err := initAudioReader(ctx, part)
				if err != nil {
					return nil, err
				}
				if res.header == nil {
					res.header = ar.audio.header
					res.bitsPerSampleV = ar.audio.bitsPerSample
					res.sampleRateV = ar.audio.sampleRate
				}
				lenBefore := res.buf.Len()
				for _, w := range part.Words {
					wwdNext.audioReader = ar
					wwdNext.word = w
					err := writeWordAudio(ctx, res, wwd, wwdNext)
					if err != nil {
						return nil, err
					}
					wwd, wwdNext = wwdNext, &wordWriteData{}
				}
				vc := makeVolumeChanges(ctx, part, lenBefore, res.buf.Len(), res.bytesPerSample())
				if len(vc) > 0 {
					volChanges = append(volChanges, vc...)
				}
			}
		}
	}
	err := writeWordAudio(ctx, res, wwd, wwdNext)
	if err != nil {
		return nil, err
	}

	if res.buf.Len() == 0 {
		return nil, errors.New("no audio")
	}
	if suffix != nil {
		if err := appendWav(ctx, res, suffix); err != nil {
			return nil, errors.Wrapf(err, "can't append suffix")
		}
	}

	resBytes, err := audio.ChangeVolume(ctx, res.buf.Bytes(), volChanges, int(res.bytesPerSample()))
	if err != nil {
		return nil, fmt.Errorf("change volume: %w", err)
	}
	var bufRes bytes.Buffer
	_, _ = bufRes.Write(res.header)
	_, _ = bufRes.Write([]byte("data"))
	_, _ = bufRes.Write(wav.SizeBytes(uint32(res.buf.Len())))
	_, _ = bufRes.Write(resBytes)

	return &synthesizer.AudioData{
		Data:          bufRes.Bytes(),
		SampleRate:    res.sampleRate(),
		BitsPerSample: res.bitsPerSample(),
	}, nil
}

func appendAudioBytes(ctx context.Context, res *wavWriter, audioReader *audioReader, toStep int) error {
	to := toStep * audioReader.step * int(audioReader.audio.bitsPerSample) / 8
	if to > len(audioReader.audio.data) {
		return errors.Errorf("append to %d > audio len %d", to, len(audioReader.audio.data))
	}
	if to < audioReader.wrote {
		return errors.Errorf("to %d < wrote %d", to, audioReader.wrote)
	}
	bData := audioReader.audio.data[audioReader.wrote:to]
	audioReader.wrote = to
	audioReader.wrotePos = toStep

	_, err := res.buf.Write(bData)
	if err != nil {
		return err
	}
	return nil
}

func initAudioReader(ctx context.Context, part *synthesizer.TTSDataPart) (*audioReader, error) {
	decoded := part.Audio
	if !wav.IsValid(decoded) {
		return nil, errors.New("no valid audio wave data")
	}
	header := wav.TakeHeader(decoded)
	parsed := &parsedWAV{
		header:        header,
		data:          wav.TakeData(decoded),
		sampleRate:    wav.GetSampleRate(header),
		bitsPerSample: wav.GetBitsPerSample(header),
	}
	es, esf := getEndSilSize(ctx, part.TranscribedSymbols, part.Durations)
	return &audioReader{
		part:        part,
		audio:       parsed,
		step:        part.Step,
		startSil:    getStartSilSize(part.TranscribedSymbols, part.Durations),
		endSil:      es,
		endSilStart: esf,
	}, nil
}

func makeVolumeChanges(ctx context.Context, part *synthesizer.TTSDataPart, startPos, endPos int, bytesPerSample uint16) []*audio.VolChange {
	var res []*audio.VolChange
	rate := part.LoudnessGain
	var last *audio.VolChange
	for _, w := range part.Words {
		if !w.Tagged.IsWord() || w.AudioPos == nil {
			continue
		}
		from := w.AudioPos.From
		for i, vc := range w.SynthesizedPos.VolumeChanges {
			d := w.SynthesizedPos.Durations[i]
			to := from + d*part.Step*int(bytesPerSample)

			gain := vc + part.LoudnessGain

			if utils.Float64Equals(gain, rate) {
				if last != nil {
					last.To = to
				}
			} else if utils.Float64Equals(gain, 0) {
				last = nil
			} else {
				v := &audio.VolChange{
					From: from,
					To:   to,
					Rate: calcVolumeRate(gain),
				}
				log.Ctx(ctx).Debug().Float64("change", calcVolumeRate(gain)).Int("from", v.From).Int("to", v.To).Msg("volume")
				res = append(res, v)
				last = v
			}
			from = to
			rate = gain
		}
	}
	if !utils.Float64Equals(part.LoudnessGain, 0) {
		// fill from startPos to endPos with part.LoudnessGain
		gainMultiply := calcVolumeRate(part.LoudnessGain)
		partRes := res
		start := startPos
		res = make([]*audio.VolChange, 0, len(partRes))
		for _, v := range partRes {
			if v.From > start {
				res = append(res, &audio.VolChange{
					From: start,
					To:   v.From,
					Rate: gainMultiply,
				})
			}
			res = append(res, v)
			start = v.To
		}
		if endPos > start {
			res = append(res, &audio.VolChange{
				From: start,
				To:   endPos,
				Rate: gainMultiply,
			})
		}
	}
	return fixStartEndRates(res, calcVolumeRate(part.LoudnessGain))
}

func fixStartEndRates(res []*audio.VolChange, defaultGain float64) []*audio.VolChange {
	var prev *audio.VolChange
	for _, v := range res {
		if prev == nil {
			v.StartRate = defaultGain
		} else if prev.To < v.From {
			prev.EndRate = defaultGain
			v.StartRate = defaultGain
		} else {
			prev.EndRate = prev.Rate
			v.StartRate = prev.Rate
		}
		prev = v
	}
	if prev != nil {
		prev.EndRate = defaultGain
	}
	return res
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
	_, err := writePause(ctx, &res.buf, res.sampleRate(), res.bitsPerSample(), pause)
	if err != nil {
		return err
	}
	// res.size += c
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
