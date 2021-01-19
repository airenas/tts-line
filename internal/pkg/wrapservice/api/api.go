package api

//Input is synthesizer input data
type Input struct {
	Text string `json:"text,omitempty"`
}

//Result is synthesis result
type Result struct {
	Data string `json:"data,omitempty"`
}
