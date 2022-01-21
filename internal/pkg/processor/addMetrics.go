package processor

import (
	"unicode/utf8"

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

//Process main processor method
func (p *addMetrics) Process(data *synthesizer.TTSData) error {
	p.mFunc(data)
	return nil
}

func getChars(data *synthesizer.TTSData) float64 {
	return float64(utf8.RuneCountInString(data.OriginalText))
}

//NewMetricsCharsFunc creates func for adding symbols count
func NewMetricsCharsFunc(url string) func(data *synthesizer.TTSData) {
	return func(data *synthesizer.TTSData) {
		totalCharMetrics.WithLabelValues(url).Add(getChars(data))
	}
}

//NewMetricsWaveLenFunc creates func for add audiolen metric
func NewMetricsWaveLenFunc(url string) func(data *synthesizer.TTSData) {
	return func(data *synthesizer.TTSData) {
		totalDurationMetrics.WithLabelValues(url).Add(data.AudioLenSeconds)
	}
}
