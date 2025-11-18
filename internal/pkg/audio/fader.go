package audio

import (
	"math"
	"time"
)

type Fader struct {
	sigmoid []float64
}

var defaultFader *Fader

func init() {
	defaultFader = newFader(50*time.Millisecond, 22050)
}

func newFader(duration time.Duration, freq uint) *Fader {
	points := float64(duration.Milliseconds()) * float64(freq) / 1000.0
	sigmoid := make([]float64, int(points))
	for i := range sigmoid {
		sigmoid[i] = calcSigmoid(float64(i)/points, 10)
	}

	return &Fader{
		sigmoid: sigmoid,
	}
}

func (f *Fader) Fade(pos int, bl int) (float64, bool /*start*/) {
	posEnd := bl - pos - 1
	if pos <= posEnd {
		return f.at(pos), true
	}
	return f.at(posEnd), false
}

func (f *Fader) at(pos int) float64 {
	if pos < 0 {
		return 0.0
	}
	if pos >= len(f.sigmoid) {
		return 1.0
	}
	return f.sigmoid[pos]
}

func calcSigmoid(x, k float64) float64 {
	return 1 / (1 + math.Exp(-k*(x-0.5)))
}
