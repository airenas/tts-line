package processor

import (
	"context"
	"fmt"

	ebur128 "git.gammaspectra.live/S.O.N.G/go-ebur128"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type calcLoudness struct {
}

// NewCalcLoudness creates loudness calculator processor
func NewCalcLoudness() synthesizer.Processor {
	return &calcLoudness{}
}

func (p *calcLoudness) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "calcLoudness.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}

	var err error
	var first float64
	for _, dp := range data.Parts {
		dp.Loudness, err = calculateLoudness(ctx, dp.Audio)
		if err != nil {
			return fmt.Errorf("calc loudness: %w", err)
		}
		log.Ctx(ctx).Debug().Float64("loudness", dp.Loudness).Msg("Part loudness")
		if utils.Float64Equals(first, 0) {
			first = dp.Loudness
		} else {
			dp.LoudnessGain = first - dp.Loudness
			log.Ctx(ctx).Debug().Float64("gain", dp.LoudnessGain).Msg("Adjusting loudness")
		}
	}

	return nil
}

type calcLoudnessSSML struct {
}

// NewCalcLoudnessSSML creates loudness calculator processor
func NewCalcLoudnessSSML() synthesizer.Processor {
	return &calcLoudnessSSML{}
}

func (p *calcLoudnessSSML) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "calcLoudness.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}

	var err error
	var first float64
	for _, dp := range data.SSMLParts {
		for _, p := range dp.Parts {
			p.Loudness, err = calculateLoudness(ctx, p.Audio)
			if err != nil {
				return fmt.Errorf("calc loudness: %w", err)
			}
			log.Ctx(ctx).Info().Float64("loudness", p.Loudness).Msg("Part loudness")
			if utils.Float64Equals(first, 0) {
				first = p.Loudness
			} else {
				p.LoudnessGain = first - p.Loudness
				log.Ctx(ctx).Info().Float64("gain", p.LoudnessGain).Msg("Adjusting loudness")
			}
		}
	}

	return nil
}

func calculateLoudness(ctx context.Context, wavData []byte) (float64, error) {
	_, span := utils.StartSpan(ctx, "calcLoudness.Process")
	defer span.End()

	if !wav.IsValid(wavData) {
		return 0, errors.New("no valid audio wave data")
	}
	header := wav.TakeHeader(wavData)
	if header == nil {
		return 0, errors.New("no valid wave header")
	}
	nch := int(wav.GetChannels(header))
	sr := int(wav.GetSampleRate(header))
	bytesPerSample := int(wav.GetBitsPerSample(header)) / 8
	if bytesPerSample != 2 {
		return 0, fmt.Errorf("unsupported bytes per sample %d", bytesPerSample)
	}

	data := wav.TakeData(wavData)
	audioInt16 := make([]int16, len(data)/2)
	for i := 0; i < len(audioInt16); i++ {
		audioInt16[i] = int16(data[2*i]) | int16(data[2*i+1])<<8
	}

	st := ebur128.NewState(nch, sr, ebur128.LoudnessGlobalMomentary)
	defer st.Close()

	err := st.AddShort(audioInt16)
	if err != nil {
		return 0, fmt.Errorf("add short: %w", err)
	}

	loud, err := st.GetLoudnessGlobal()
	if err != nil {
		return 0, fmt.Errorf("get loudness: %w", err)
	}
	return loud, nil
}

// Info return info about processor
func (p *calcLoudness) Info() string {
	return "calcLoudness()"
}

// Info return info about processor
func (p *calcLoudnessSSML) Info() string {
	return "calcLoudness()"
}
