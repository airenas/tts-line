package api

import "github.com/airenas/tts-line/internal/pkg/syntmodel"

// Params synthesis configuration params
type Params struct {
	Text            string
	Speed           float32
	Voice           string
	Priority        int
	DurationsChange []float64
	PitchChange     [][]*syntmodel.PitchChange
}
