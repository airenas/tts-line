package processor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	ebur128 "git.gammaspectra.live/S.O.N.G/go-ebur128"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/rs/zerolog/log"
)

type calcLoudness struct {
	parallelWorkers int
}

// NewCalcLoudness creates loudness calculator processor
func NewCalcLoudness(parallelWorkers int) synthesizer.Processor {
	if parallelWorkers < 1 {
		parallelWorkers = 3
	}
	log.Info().Int("workers", parallelWorkers).Msg("Loudness calculator")
	return &calcLoudness{parallelWorkers: parallelWorkers}
}

func (p *calcLoudness) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "calcLoudness.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}

	return addLoudness(ctx, data.Parts, p.parallelWorkers)
}

func addLoudness(ctx context.Context, parts []*synthesizer.TTSDataPart, parallelWorkers int) error {
	if len(parts) < 2 { // nothing to adjust
		return nil
	}

	if err := calculatePartsLoudness(ctx, parts, parallelWorkers); err != nil {
		return fmt.Errorf("calculate parts loudness: %w", err)
	}

	if err := adjustPartsLoudness(ctx, parts); err != nil {
		return fmt.Errorf("adjust parts loudness: %w", err)
	}
	return nil
}

func calculatePartsLoudness(ctx context.Context, parts []*synthesizer.TTSDataPart, parallelWorkers int) error {
	ctx, span := utils.StartSpan(ctx, "calculatePartsLoudness")
	defer span.End()

	workerQueueLimit := make(chan bool, parallelWorkers)
	errCh := make(chan error, 1)
	closeCh := make(chan bool, 1)
	defer close(closeCh)

	var wg sync.WaitGroup

	for _, part := range parts {
		select {
		case err := <-errCh:
			return fmt.Errorf("process partial calculateLoudness tasks: %w", err)
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		case workerQueueLimit <- true:
			wg.Add(1)
			go func(part *synthesizer.TTSDataPart) {
				defer wg.Done()
				defer func() { <-workerQueueLimit }()

				loudness, err := calculateLoudness(ctx, part.Audio)
				if err != nil {
					select {
					case <-closeCh:
					case errCh <- err:
					}
				} else {
					part.Loudness = loudness
					log.Ctx(ctx).Debug().Float64("loudness", loudness).Msg("Calculated loudness")
				}
			}(part)
		}
	}

	waitCh := make(chan bool, 1)
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("process partial calculateLoudness tasks: %w", err)
	case <-waitCh:
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
	return nil
}

func adjustPartsLoudness(ctx context.Context, parts []*synthesizer.TTSDataPart) error {
	target := 0.0
	for _, p := range parts {
		if hasNormalPAFSValue(p.Loudness) {
			target = p.Loudness
			break
		}
	}
	if !hasNormalPAFSValue(target) {
		log.Ctx(ctx).Warn().Msg("No normal loudness found, skip adjustment")
		return nil
	}
	for _, p := range parts {
		if hasNormalPAFSValue(p.Loudness) {
			p.LoudnessGain = target - p.Loudness
			log.Ctx(ctx).Debug().Float64("gain", p.LoudnessGain).Msg("Adjusting loudness")
		}
	}
	return nil
}

func hasNormalPAFSValue(first float64) bool {
	return first < -10 && first > -30 // normal loudness range
}

type calcLoudnessSSML struct {
	parallelWorkers int
}

// NewCalcLoudnessSSML creates loudness calculator processor
func NewCalcLoudnessSSML(parallelWorkers int) synthesizer.Processor {
	if parallelWorkers < 1 {
		parallelWorkers = 3
	}
	log.Info().Int("workers", parallelWorkers).Msg("Loudness SSML calculator")
	return &calcLoudnessSSML{parallelWorkers: parallelWorkers}
}

func (p *calcLoudnessSSML) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "calcLoudnessSSML.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}

	return addLoudness(ctx, collectPartsFromSSML(data.SSMLParts), p.parallelWorkers)
}

func collectPartsFromSSML(ssmlParts []*synthesizer.TTSData) []*synthesizer.TTSDataPart {
	var res []*synthesizer.TTSDataPart
	for _, dp := range ssmlParts {
		res = append(res, dp.Parts...)
	}
	return res
}

func calculateLoudness(ctx context.Context, wavData []byte) (float64, error) {
	_, span := utils.StartSpan(ctx, "calcLoudness.calculateLoudness")
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
	if loud < -70 { // too quiet
		loud = -70
	}
	return loud, nil
}

// Info return info about processor
func (p *calcLoudness) Info() string {
	return "calcLoudness()"
}

// Info return info about processor
func (p *calcLoudnessSSML) Info() string {
	return "calcLoudnessSSML()"
}
