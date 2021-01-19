package api

//Input is synthesizer input data
type Input struct {
	Text string `json:"text,omitempty"`
}

//Result is synthesis result
type Result struct {
	AudioAsString string `json:"audioAsString,omitempty"`
}
