package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewAddMetrics(t *testing.T) {
	pr, err := NewAddMetrics(nil)
	assert.Nil(t, pr)
	assert.NotNil(t, err)
	pr, err = NewAddMetrics(NewMetricsCharsFunc("test"))
	assert.NotNil(t, pr)
	assert.Nil(t, err)
	pr, err = NewAddMetrics(NewMetricsWaveLenFunc("test"))
	assert.NotNil(t, pr)
	assert.Nil(t, err)
}

func TestCallsCharMetrics(t *testing.T) {
	pr, err := NewAddMetrics(NewMetricsCharsFunc("test"))
	assert.NotNil(t, pr)
	assert.Nil(t, err)
	d := &synthesizer.TTSData{}
	d.OriginalText = "0123456789"
	err = pr.Process(d)
	assert.Nil(t, err)
	err = pr.Process(d)
	assert.Nil(t, err)
	assert.InDelta(t, 20.0, testutil.ToFloat64(totalCharMetrics.WithLabelValues("test")), 0.000001)
}

func TestCallsCMetricsWaveLen(t *testing.T) {
	pr, err := NewAddMetrics(NewMetricsWaveLenFunc("test"))
	assert.NotNil(t, pr)
	assert.Nil(t, err)
	d := &synthesizer.TTSData{}
	d.AudioDuration = 0.35
	err = pr.Process(d)
	assert.Nil(t, err)
	err = pr.Process(d)
	assert.Nil(t, err)
	err = pr.Process(d)
	assert.Nil(t, err)
	assert.InDelta(t, 1.05, testutil.ToFloat64(totalDurationMetrics.WithLabelValues("test")), 0.000001)
}
