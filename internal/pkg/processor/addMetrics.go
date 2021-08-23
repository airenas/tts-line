package processor

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

var totalCharMetrics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tts_input_chars_total",
		Help: "The total number of inputed chars",
	},
	[]string{"url"},
)

var totalDurationMetrics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tts_output_wave_len_seconds_total",
		Help: "The total number of outputed audio len",
	},
	[]string{"url"},
)

func init() {
	prometheus.MustRegister(totalCharMetrics, totalDurationMetrics)
}

type addMetrics struct {
	mFunc func(data *synthesizer.TTSData)
}

//NewAddMetrics creates new processor to fill metrics
func NewAddMetrics(mFunc func(data *synthesizer.TTSData)) (synthesizer.Processor, error) {
	if mFunc == nil {
		return nil, errors.New("no metric function")
	}
	return &addMetrics{mFunc: mFunc}, nil
}

func (p *addMetrics) Process(data *synthesizer.TTSData) error {
	p.mFunc(data)
	return nil
}

func getChars(data *synthesizer.TTSData) float64 {
	return float64(len([]rune(data.OriginalText)))
}

func NewMetricsCharsFunc(url string) func(data *synthesizer.TTSData) {
	return func(data *synthesizer.TTSData) {
		totalCharMetrics.WithLabelValues(url).Add(getChars(data))
	}
}

func NewMetricsWaveLenFunc(url string) func(data *synthesizer.TTSData) {
	return func(data *synthesizer.TTSData) {
		totalDurationMetrics.WithLabelValues(url).Add(data.AudioLenSeconds)
	}
}
