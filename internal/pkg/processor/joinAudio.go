package processor

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/pkg/errors"
)

type joinAudio struct {
}

//NewJoinAudio joins results into one audio
func NewJoinAudio() synthesizer.Processor {
	return &joinAudio{}
}

func (p *joinAudio) Process(data *synthesizer.TTSData) error {
	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	var err error
	data.Audio, err = join(data.Parts)
	if err != nil {
		return errors.Wrap(err, "Can't join audio")
	}
	utils.LogData("Output: ", fmt.Sprintf("audio len %d", len(data.Audio)))
	return nil
}

func join(parts []*synthesizer.TTSDataPart) (string, error) {
	var buf bytes.Buffer
	var bufHeader []byte
	size := uint32(0)

	for _, part := range parts {
		decoded, err := base64.StdEncoding.DecodeString(part.Audio)
		if err != nil {
			return "", err
		}
		if !wav.IsValid(decoded) {
			return "", errors.New("No valid audio wave data")
		}
		if bufHeader == nil {
			bufHeader = wav.TakeHeader(decoded)
		}

		size += wav.GetSize(decoded)
		_, err = buf.Write(wav.TakeData(decoded))
		if err != nil {
			return "", err
		}
	}
	var bufRes bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &bufRes)
	enc.Write(bufHeader)
	enc.Write(wav.SizeBytes(size))
	enc.Write(buf.Bytes())
	err := enc.Close()
	if err != nil {
		return "", err
	}
	return bufRes.String(), nil
}
