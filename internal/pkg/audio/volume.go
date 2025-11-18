package audio

import (
	"fmt"
	"math"
)

type VolChange struct {
	From int // pos in bytes
	To   int // pos in bytes

	Rate      float64
	StartRate float64
	EndRate   float64
}

func ChangeVolume(b []byte, volChange []*VolChange, bytesPerSaample int) ([]byte, error) {
	if len(volChange) == 0 {
		return b, nil
	}
	for _, vc := range volChange {
		for i := vc.From; i < vc.To; i += bytesPerSaample {
			l := (vc.To - vc.From) / bytesPerSaample
			switch bytesPerSaample {
			case 2:
				fade, isStart := defaultFader.Fade((i-vc.From)/bytesPerSaample, l)
				sample := int16(b[i]) | int16(b[i+1])<<8
				rate := 1.0
				if isStart {
					rate = vc.StartRate + (vc.Rate-vc.StartRate)*fade
				} else {
					rate = vc.EndRate + (vc.Rate-vc.EndRate)*(fade)
				}
 				newSample := toInt16(float64(sample) * rate)
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
