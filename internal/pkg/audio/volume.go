package audio

import (
	"context"
	"fmt"
	"math"

	"github.com/airenas/tts-line/internal/pkg/utils"
)

type VolChange struct {
	From int // pos in bytes
	To   int // pos in bytes

	Rate      float64
	StartRate float64
	EndRate   float64
}

func ChangeVolume(ctx context.Context, b []byte, volChange []*VolChange, bytesPerSample int) ([]byte, error) {
	if len(volChange) == 0 {
		return b, nil
	}

	_, span := utils.StartSpan(ctx, "audio.ChangeVolume")
	defer span.End()

	lb := len(b)
	for _, vc := range volChange {
		l := (vc.To - vc.From) / bytesPerSample
		for i := vc.From; i < vc.To; i += bytesPerSample {
			if (i + bytesPerSample) > lb {
				return nil, fmt.Errorf("out of bounds volume change: %d + %d > %d", i, bytesPerSample, lb)
			}
			switch bytesPerSample {
			case 2:
				fade, isStart := defaultFader.Fade((i-vc.From)/bytesPerSample, l)
				sample := int16(b[i]) | int16(b[i+1])<<8
				var rate float64
				if isStart {
					rate = vc.StartRate + (vc.Rate-vc.StartRate)*fade
				} else {
					rate = vc.EndRate + (vc.Rate-vc.EndRate)*(fade)
				}
				newSample := toInt16(float64(sample) * rate)
				b[i] = byte(newSample & 0xFF)
				b[i+1] = byte((newSample >> 8) & 0xFF)
			default:
				return nil, fmt.Errorf("unsupported bytes per sample %d", bytesPerSample)
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
