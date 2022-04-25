package api

//Input is sythesis input data
type Input struct {
	Text string `json:"text,omitempty"`
	//Possible values are m4a, mp3
	OutputFormat     string  `json:"outputFormat,omitempty"`
	OutputTextFormat string  `json:"outputTextFormat,omitempty"`
	AllowCollectData *bool   `json:"saveRequest,omitempty"`
	Speed            float32 `json:"speed,omitempty"`
	Voice            string  `json:"voice,omitempty"`
	Priority         int     `json:"priority,omitempty"`
}

//Result is synthesis result
type Result struct {
	AudioAsString string `json:"audioAsString,omitempty"`
	Error         string `json:"error,omitempty"`
	Text          string `json:"text,omitempty"`
	RequestID     string `json:"requestID,omitempty"`
}

//InfoResult is a responce for /synthesizeInfo request
type InfoResult struct {
	Count int64 `json:"count"`
}
