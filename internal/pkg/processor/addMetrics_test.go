package processor

import (
	"context"
	"testing"
	"time"

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
	err = pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	err = pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.InDelta(t, 20.0, testutil.ToFloat64(totalCharMetrics.WithLabelValues("test")), 0.000001)
}

func TestCallsCMetricsWaveLen(t *testing.T) {
	pr, err := NewAddMetrics(NewMetricsWaveLenFunc("test"))
	assert.NotNil(t, pr)
	assert.Nil(t, err)
	d := &synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, make([]byte, 22050*2*.35)) // 0.35 sec of audio
	err = pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	err = pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	err = pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.InDelta(t, 1.05, testutil.ToFloat64(totalDurationMetrics.WithLabelValues("test")), 0.000001)
}

func testGenerateSampleData(t *testing.T, data []byte) *synthesizer.AudioData {
	t.Helper()

	return &synthesizer.AudioData{Data: data, SampleRate: 22050, BitsPerSample: 16, Duration: time.Duration(len(data)) * time.Second / time.Duration(22050*2)}
}
