package api

const (
	SpeechMarkTypeWord = "word"
)

// Input is sythesis input data
type Input struct {
	Text string `json:"text,omitempty"`
	//TextType may have values: text, ssml
	TextType string `json:"textType,omitempty"`
	//Possible values are m4a, mp3
	OutputFormat     string  `json:"outputFormat,omitempty"`
	OutputTextFormat string  `json:"outputTextFormat,omitempty"`
	AllowCollectData *bool   `json:"saveRequest,omitempty"`
	Speed            float32 `json:"speed,omitempty"`
	Voice            string  `json:"voice,omitempty"`
	Priority         int     `json:"priority,omitempty"`
	//Possible values are: word
	SpeechMarkTypes []string `json:"speechMarkTypes,omitempty"`
}

// SpeechMark
type SpeechMark struct {
	//In Millis from the start of the audio
	TimeInMillis int64 `json:"timeMillis,omitempty"`
	//In Millis
	Duration int64 `json:"durationMillis,omitempty"`
	//Possible values are: word
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Result is synthesis result
type Result struct {
	AudioAsString string        `json:"audioAsString,omitempty"`
	Error         string        `json:"error,omitempty"`
	Text          string        `json:"text,omitempty"`
	RequestID     string        `json:"requestID,omitempty"`
	SpeechMarks   []*SpeechMark `json:"speechMarks,omitempty"`
}

// InfoResult is a response for /synthesizeInfo request
type InfoResult struct {
	Count int64 `json:"count"`
}
