package api

//Input is sythesis input data
type Input struct {
	Text string `json:"text,omitempty"`
}

type check struct {
	String string `json:"text,omitempty"`
	ID     int    `json:"manual,omitempty"`
}

type failure struct {
	FailingText     string `json:"text,omitempty"`
	FailingPosition int    `json:"manual,omitempty"`
	Check           check  `json:"validTo"`
}

//Result is synthesis result
type Result struct {
	AudioAsString      string    `json:"audioAsString,omitempty"`
	Error              string    `json:"error,omitempty"`
	ValidationFailures []failure `json:"validationFailures,omitempty"`
}
