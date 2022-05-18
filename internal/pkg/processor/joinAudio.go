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

type joinAudio struct {
}

// NewJoinAudio joins results into one audio
func NewJoinAudio() synthesizer.Processor {
	return &joinAudio{}
}

func (p *joinAudio) Process(data *synthesizer.TTSData) error {
	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var err error
	data.Audio, data.AudioLenSeconds, err = join(data.Parts)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

func join(parts []*synthesizer.TTSDataPart) (string, float64, error) {
	var buf bytes.Buffer
	var bufHeader []byte
	size := uint32(0)

	for _, part := range parts {
		decoded, err := base64.StdEncoding.DecodeString(part.Audio)
		if err != nil {
			return "", 0, err
		}
		if !wav.IsValid(decoded) {
			return "", 0, errors.New("No valid audio wave data")
		}
		if bufHeader == nil {
			bufHeader = wav.TakeHeader(decoded)
		}

		size += wav.GetSize(decoded)
		_, err = buf.Write(wav.TakeData(decoded))
		if err != nil {
			return "", 0, err
		}
	}
	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	enc.Write(bufHeader)
	enc.Write(wav.SizeBytes(size))
	enc.Write(buf.Bytes())
	err := enc.Close()
	if err != nil {
		return "", 0, err
	}
	bitsRate := wav.GetBitsRateCalc(bufHeader)
	if bitsRate == 0 {
		return "", 0, errors.New("can't extract bits rate from header")
	}
	return bufRes.String(), float64(size) / float64(bitsRate), nil
}

type joinSSMLAudio struct {
}

//NewJoinSSMLAudio joins results into one audio from many ssml parts
func NewJoinSSMLAudio() synthesizer.Processor {
	return &joinSSMLAudio{}
}

func (p *joinSSMLAudio) Process(data *synthesizer.TTSData) error {
	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var err error
	data.Audio, data.AudioLenSeconds, err = joinSSML(data)
	if err != nil {
		return errors.Wrap(err, "can't join audio")
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

func joinSSML(data *synthesizer.TTSData) (string, float64, error) {
	var buf bytes.Buffer
	var bufHeader []byte
	size := uint32(0)
	pause := time.Duration(0)
	bitsRate := uint32(0)

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
				if bufHeader == nil {
					bufHeader = wav.TakeHeader(decoded)
					bitsRate = wav.GetBitsRateCalc(bufHeader)
				}

				if pause > 0 {
					s, err := writePause(buf, bitsRate, pause)
					if err != nil {
						return "", 0, err
					}
					size += s
					pause = 0
				}
				size += wav.GetSize(decoded)
				_, err = buf.Write(wav.TakeData(decoded))
				if err != nil {
					return "", 0, err
				}
			}
		}
	}
	if bitsRate == 0 {
		return "", 0, errors.Errorf("no audio")
	}

	if pause > 0 {
		s, err := writePause(buf, bitsRate, pause)
		if err != nil {
			return "", 0, err
		}
		size += s
	}

	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	enc.Write(bufHeader)
	enc.Write(wav.SizeBytes(size))
	enc.Write(buf.Bytes())
	err := enc.Close()
	if err != nil {
		return "", 0, err
	}
	return bufRes.String(), float64(size) / float64(bitsRate), nil
}

func writePause(buf bytes.Buffer, bitsRate uint32, pause time.Duration) (uint32, error) {
	if pause > time.Second*10 {
		goapp.Log.Warnf("Too long pause %v", pause)
		pause = time.Second * 10
	}
	if pause < 0 {
		pause = 0
	}
	c := uint32(pause.Milliseconds() * int64(bitsRate) / 1000)
	for i := uint32(0); i < c; i++ {
		if err := buf.WriteByte(0); err != nil {
			return 0, err
		}
	}
	return c, nil
}
