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
		for i := vc.From; i < vc.To-bytesPerSample; i += bytesPerSample {
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

func SearchBestSilAround(ctx context.Context, b []byte, bytesPerSample uint16, center, window int) int {
	if bytesPerSample != 2 {
		return center
	}
	lb := len(b) / int(bytesPerSample)

	start := center - window
	if start < 0 {
		start = 0
	}
	end := center + window - 1
	if end > lb {
		end = lb - 1
	}

	sw := 22050 / 100 // 10ms
	samples := make([]int16, sw)
	if end-start < sw {
		return center
	}

	sum, swPos, i := 0, 0, start
	for ; i < start+sw; i++ {
		samples[swPos] = int16(b[i*2]) | int16(b[i*2+1])<<8
		sum += abs(samples[swPos])
		swPos++
	}

	bestS, bestP := sum, i-sw/2
	swPos = 0
	for ; i < end; i++ {
		sum -= abs(samples[swPos])
		// add new
		samples[swPos] = int16(b[i*2]) | int16(b[i*2+1])<<8
		sum += abs(samples[swPos])

		if sum < bestS {
			bestS = sum
			bestP = i - sw/2
		}
		swPos = (swPos + 1) % sw
	}
	return bestP
}

func abs(i int16) int {
	if i < 0 {
		return -int(i)
	}
	return int(i)
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
