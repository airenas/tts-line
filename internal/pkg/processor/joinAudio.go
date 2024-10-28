package processor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/pkg/errors"
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

func (p *joinAudio) Process(data *synthesizer.TTSData) error {
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

	data.Audio, data.AudioLenSeconds, data.SampleRate, err = join(data.Parts, suffix)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

type wavWriter struct {
	header []byte
	size   uint32
	buf    bytes.Buffer
}

func (wr *wavWriter) SampleRate() uint32 {
	if wr.header == nil {
		return 0
	}
	return wav.GetSampleRate(wr.header)
}

func (wr *wavWriter) BitsPerSample() uint16 {
	if wr.header == nil {
		return 0
	}
	return wav.GetBitsPerSample(wr.header)
}

func join(parts []*synthesizer.TTSDataPart, suffix []byte) (string, float64, uint32 /*sampleRate*/, error) {
	res := &wavWriter{}
	nextStartSil := 0
	for i, part := range parts {
		decoded, err := base64.StdEncoding.DecodeString(part.Audio)
		if err != nil {
			return "", 0, 0, err
		}
		startSil, endSil := nextStartSil, 0
		if i == 0 {
			startSil = getStartSilSize(part.TranscribedSymbols, part.Durations)
			startSil, _, _ = calcPauseWithEnds(startSil, 0, part.DefaultSilence/2)
		}
		if i == len(parts)-1 {
			endSil = getEndSilSize(part.TranscribedSymbols, part.Durations)
			if suffix != nil {
				_, endSil, _ = calcPauseWithEnds(0, endSil, part.DefaultSilence)
			} else {
				_, endSil, _ = calcPauseWithEnds(0, endSil, part.DefaultSilence/2)
			}
		} else if i < len(parts)-1 {
			endSil = getEndSilSize(part.TranscribedSymbols, part.Durations)
			nextStartSil = getStartSilSize(parts[i+1].TranscribedSymbols, parts[i+1].Durations)
			endSil, nextStartSil, _ = calcPauseWithEnds(endSil, nextStartSil, part.DefaultSilence)
		}
		startSkip, endSkip := startSil*part.Step, endSil*part.Step
		lenBefore := res.buf.Len()
		if err := appendWav(res, decoded, startSkip, endSkip); err != nil {
			return "", 0, 0, err
		}
		part.AudioDurations = &synthesizer.AudioDurations{
			Shift:    -startSil,
			Duration: calculateDurations(res.buf.Len()-lenBefore, res.SampleRate()*uint32(res.BitsPerSample())/uint32(8)),
		}
	}
	if res.size == 0 {
		return "", 0, 0, errors.New("no wav data")
	}
	if suffix != nil {
		if err := appendWav(res, suffix, 0, 0); err != nil {
			return "", 0, 0, errors.Wrapf(err, "can't append suffix")
		}
	}
	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	_, _ = enc.Write(res.header)

	_, _ = enc.Write([]byte("data"))
	_, _ = enc.Write(wav.SizeBytes(res.size))
	_, _ = enc.Write(res.buf.Bytes())
	if err := enc.Close(); err != nil {
		return "", 0, 0, err
	}
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return "", 0, 0, errors.New("can't extract bits rate from header")
	}
	os.WriteFile("testout..wav.base64", bufRes.Bytes(), 0644)
	return bufRes.String(), float64(res.size) / float64(bitsRate), wav.GetSampleRate(res.header), nil
}

func calculateDurations(aLen int, samplesPerSec uint32) time.Duration {
	if samplesPerSec == 0 {
		return 0
	}
	return time.Duration(float64(aLen) * 1000/float64(samplesPerSec)) * time.Millisecond
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

func getEndSilSize(phones []string, durations []int) int {
	l := len(phones)
	if len(durations) != l+1 {
		goapp.Log.Warn("Duration size don't match phone list")
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
	os.WriteFile("test.wav", wavData, 0644)
	header := wav.TakeHeader(wavData)
	if res.header == nil {
		res.header = header
	} else {
		if wav.GetSampleRate(res.header) != wav.GetSampleRate(header) {
			return errors.Errorf("differs sample rate %d vs %d", wav.GetSampleRate(res.header), wav.GetSampleRate(header))
		}
		if wav.GetBitsPerSample(res.header) != wav.GetBitsPerSample(header) {
			return errors.Errorf("differs bits per sample %d vs %d", wav.GetBitsPerSample(res.header), wav.GetBitsPerSample(header))
		}
	}
	bData := wav.TakeData(wavData)
	bps := res.BitsPerSample()
	fix := 1
	if bps != 0 {
		fix = int(bps / 8)
	}
	es := len(bData) - (endSkip * fix)
	if es < (startSkip * fix) {
		return errors.Errorf("audio start pos > end: %d vs %d", startSkip, es)
	}

	res.size += uint32(es - (startSkip * fix)) // wav.GetSize(wavData)

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

func (p *joinSSMLAudio) Process(data *synthesizer.TTSData) error {
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
	data.Audio, data.AudioLenSeconds, data.SampleRate, err = joinSSML(data, suffix)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	for _, dp := range data.SSMLParts {
		dp.SampleRate = data.SampleRate
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

type nextWriteData struct {
	part        *synthesizer.TTSDataPart
	data        []byte
	startSkip   int
	pause       time.Duration // pause that should be written after data
	isPause     bool
	durationAdd time.Duration // time to add to the part
}

func joinSSML(data *synthesizer.TTSData, suffix []byte) (string, float64 /*sampleRate*/, uint32, error) {
	res := &wavWriter{}
	wd := &nextWriteData{}
	wd.pause = time.Duration(0)
	writeF := func(part *synthesizer.TTSDataPart) error {
		step, defaultSil, pause := 0, 0, time.Duration(0)
		endSil, startSil := 0, 0
		var decoded []byte
		var err error
		if part != nil {
			decoded, err = base64.StdEncoding.DecodeString(part.Audio)
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
			// endSil = getEndSilSize(part.TranscribedSymbols, part.Durations)
			step = part.Step
			defaultSil = part.DefaultSilence
		}
		if wd.part != nil {
			endSil = getEndSilSize(wd.part.TranscribedSymbols, wd.part.Durations)
			// startSil = getStartSilSize(wd.part.TranscribedSymbols, wd.part.Durations)
			step = wd.part.Step
			defaultSil = wd.part.DefaultSilence
		}
		pause = 0
		if wd.isPause {
			startSil, endSil, pause = fixPause(startSil, endSil, wd.pause, step)
		} else {
			startSil, endSil, _ = calcPauseWithEnds(startSil, endSil, defaultSil)
		}

		startSkip, endSkip := startSil*step, endSil*step
		if wd.part != nil {
			lenBefore := res.buf.Len()
			if err := appendWav(res, wd.data, wd.startSkip, endSkip); err != nil {
				return err
			}
			wd.part.AudioDurations = &synthesizer.AudioDurations{
				Shift:    -wd.startSkip / step,
				Duration: calculateDurations(res.buf.Len()-lenBefore, res.SampleRate()*uint32(res.BitsPerSample())/uint32(8)),
			}
			wd.part.AudioDurations.Shift += utils.ToTTSSteps(wd.durationAdd, wav.GetSampleRate(res.header), step)
			wd.part.AudioDurations.Duration += wd.durationAdd
			wd.durationAdd = 0
		}
		if pause > 0 {
			if err := appendPause(res, pause); err != nil {
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
				err := writeF(part)
				if err != nil {
					return "", 0, 0, err
				}
			}
		}
	}
	if err := writeF(nil); err != nil {
		return "", 0, 0, err
	}
	if res.size == 0 {
		return "", 0, 0, errors.New("no audio")
	}
	if suffix != nil {
		if err := appendWav(res, suffix, wd.startSkip, 0); err != nil {
			return "", 0, 0, errors.Wrapf(err, "can't append suffix")
		}
	}

	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	_, _ = enc.Write(res.header)

	_, _ = enc.Write([]byte("data"))
	_, _ = enc.Write(wav.SizeBytes(res.size))
	_, _ = enc.Write(res.buf.Bytes())

	if err := enc.Close(); err != nil {
		return "", 0, 0, err
	}
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return "", 0, 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.String(), float64(res.size) / float64(bitsRate), wav.GetSampleRate(res.header), nil
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

func fixPause(s1, s2 int, pause time.Duration, step int) (int, int, time.Duration) {
	if step == 0 {
		return s1, s2, pause
	}
	millisInHops := float64(step) / float64(22.050)
	pauseHops := int(math.Round(float64(pause.Milliseconds()) / millisInHops))
	r1, r2, rp := calcPauseWithEnds(s1, s2, pauseHops)
	return r1, r2, time.Millisecond * time.Duration(int(float64(rp)*millisInHops))
}

func appendPause(res *wavWriter, pause time.Duration) error {
	if res.header == nil {
		return errors.New("no wav data before pause")
	}
	c, err := writePause(&res.buf, wav.GetSampleRate(res.header), wav.GetBitsPerSample(res.header), pause)
	if err != nil {
		return err
	}
	res.size += c
	return nil
}

func writePause(buf *bytes.Buffer, sampleRate uint32, bitsPerSample uint16, pause time.Duration) (uint32, error) {
	if pause > time.Second*10 {
		goapp.Log.Warnf("Too long pause %v", pause)
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
