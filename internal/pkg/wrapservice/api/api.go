package api

//Input is synthesizer input data
type Input struct {
	Text  string  `json:"text,omitempty"`
	Speed float32 `json:"speedAlpha,omitempty"`
	Voice string  `json:"voice,omitempty"`
	Priority int  `json:"priority,omitempty"`
}

//Result is synthesis result
type Result struct {
	Data string `json:"data,omitempty"`
}
