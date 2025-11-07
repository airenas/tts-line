package syntmodel

type PitchChange struct {
	Value float64 `json:"v,omitempty"`
	Type  int     `json:"t,omitempty"`
}

// Input is synthesizer input data
type AMInput struct {
	Text            string           `json:"text,omitempty"`
	Speed           float32          `json:"speedAlpha,omitempty"`
	Voice           string           `json:"voice,omitempty"`
	Priority        int              `json:"priority,omitempty"`
	DurationsChange []float64        `json:"durationsChange,omitempty"`
	PitchChange     [][]*PitchChange `json:"pitchChange,omitempty"`
}

type AMOutput struct {
	Data        []byte `json:"data,omitempty" msgpack:"data,omitempty"`
	Durations   []int  `json:"durations,omitempty" msgpack:"durations,omitempty"`
	SilDuration int    `json:"silDuration,omitempty" msgpack:"silDuration,omitempty"`
	Step        int    `json:"step,omitempty" msgpack:"step,omitempty"`
}

type VocInput struct {
	Data     []byte `json:"data" msgpack:"data"`
	Voice    string `json:"voice" msgpack:"voice"`
	Priority int    `json:"priority,omitempty" msgpack:"priority,omitempty"`
}

type VocOutput struct {
	Data []byte `json:"data" msgpack:"data"`
}

// Result is synthesis result
type Result struct {
	Data        []byte `json:"data,omitempty" msgpack:"data,omitempty"`
	Durations   []int  `json:"durations,omitempty" msgpack:"durations,omitempty"`
	SilDuration int    `json:"silDuration,omitempty" msgpack:"silDuration,omitempty"`
	Step        int    `json:"step,omitempty" msgpack:"step,omitempty"`
	Error       string `json:"error,omitempty" msgpack:"error,omitempty"`
}
