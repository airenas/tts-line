package api

//Input is sythesis input data
type Input struct {
	Text string `json:"text,omitempty"`
	//Possible values are m4a, mp3
	OutputFormat string `json:"outputFormat,omitempty"`
}

//Check is validation check
type Check struct {
	ID    string `json:"id,omitempty"`
	Value int    `json:"value,omitempty"`
}

//ValidateFailure indicate validation failure position
type ValidateFailure struct {
	FailingText     string `json:"failingText,omitempty"`
	FailingPosition int    `json:"failingPosition,omitempty"`
	Check           Check  `json:"check"`
}

//Result is synthesis result
type Result struct {
	AudioAsString      string            `json:"audioAsString,omitempty"`
	Error              string            `json:"error,omitempty"`
	ValidationFailures []ValidateFailure `json:"validationFailItems,omitempty"`
}

//TTSRequestConfig config for request
type TTSRequestConfig struct {
	Text           string
	OutputFormat   string
	OutputMetadata []string
}
