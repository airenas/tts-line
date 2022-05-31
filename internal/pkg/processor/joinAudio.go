package processor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

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
	if data.SuffixName != "" {
		if suffix, err = p.suffixProvider.TakeWav(data.SuffixName); err != nil {
			return errors.Wrapf(err, "can't take suffix %s", data.SuffixName)
		}
	}
	data.Audio, data.AudioLenSeconds, err = join(data.Parts, suffix)
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

func join(parts []*synthesizer.TTSDataPart, suffix []byte) (string, float64, error) {
	res := &wavWriter{}

	for _, part := range parts {
		decoded, err := base64.StdEncoding.DecodeString(part.Audio)
		if err != nil {
			return "", 0, err
		}
		if err := appendWav(res, decoded); err != nil {
			return "", 0, err
		}
	}
	if res.size == 0 {
		return "", 0, errors.New("no wav data")
	}
	if suffix != nil {
		if err := appendWav(res, suffix); err != nil {
			return "", 0, errors.Wrapf(err, "can't append suffix")
		}
	}
	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	enc.Write(res.header)
	enc.Write(wav.SizeBytes(res.size))
	enc.Write(res.buf.Bytes())
	if err := enc.Close(); err != nil {
		return "", 0, err
	}
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return "", 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.String(), float64(res.size) / float64(bitsRate), nil
}

func appendWav(res *wavWriter, wavData []byte) error {
	if !wav.IsValid(wavData) {
		return errors.New("no valid audio wave data")
	}
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
	res.size += wav.GetSize(wavData)
	_, err := res.buf.Write(wav.TakeData(wavData))
	return err
}

// Info return info about processor
func (p *joinAudio) Info() string {
	return fmt.Sprintf("joinAudio(%s)", retrieveInfo(p.suffixProvider))
}

type joinSSMLAudio struct {
	suffixProvider AudioLoader
}

//NewJoinSSMLAudio joins results into one audio from many ssml parts
func NewJoinSSMLAudio(suffixProvider AudioLoader) synthesizer.Processor {
	return &joinSSMLAudio{suffixProvider: suffixProvider}
}

func (p *joinSSMLAudio) Process(data *synthesizer.TTSData) error {
	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var suffix []byte
	var err error
	if data.SuffixName != "" {
		if suffix, err = p.suffixProvider.TakeWav(data.SuffixName); err != nil {
			return errors.Wrapf(err, "can't take suffix %s", data.SuffixName)
		}
	}
	data.Audio, data.AudioLenSeconds, err = joinSSML(data, suffix)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

func joinSSML(data *synthesizer.TTSData, suffix []byte) (string, float64, error) {
	res := &wavWriter{}
	pause := time.Duration(0)
	for _, dp := range data.SSMLParts {
		switch dp.Cfg.Type {
		case synthesizer.SSMLPause:
			pause = pause + dp.Cfg.PauseDuration
		case synthesizer.SSMLText:
			for _, part := range dp.Parts {
				decoded, err := base64.StdEncoding.DecodeString(part.Audio)
				if err != nil {
					return "", 0, err
				}
				if !wav.IsValid(decoded) {
					return "", 0, errors.New("no valid audio wave data")
				}
				if res.header == nil {
					res.header = wav.TakeHeader(decoded)
				}

				if pause > 0 {
					if err := appendPause(res, pause); err != nil {
						return "", 0, err
					}
					pause = 0
				}
				if err := appendWav(res, decoded); err != nil {
					return "", 0, err
				}
			}
		}
	}
	if res.size == 0 {
		return "", 0, errors.New("no audio")
	}
	if pause > 0 {
		if err := appendPause(res, pause); err != nil {
			return "", 0, err
		}
	}
	if suffix != nil {
		if err := appendWav(res, suffix); err != nil {
			return "", 0, errors.Wrapf(err, "can't append suffix")
		}
	}

	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	enc.Write(res.header)
	enc.Write(wav.SizeBytes(res.size))
	enc.Write(res.buf.Bytes())

	if err := enc.Close(); err != nil {
		return "", 0, err
	}
	bitsRate := wav.GetBitsRateCalc(res.header)
	if bitsRate == 0 {
		return "", 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.String(), float64(res.size) / float64(bitsRate), nil
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
	return fmt.Sprintf("joinSSMLAudio(%s)", retrieveInfo(p.suffixProvider))
}

func retrieveInfo (pr interface {}) string {
	pri, ok := pr.(interface {
		Info() string
	})
	if ok {
		return pri.Info()
	}
	return ""
}
