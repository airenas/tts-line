package api

const (
	SpeechMarkTypeWord = "word"
)

type SymbolMode string

const (
	SymbolModeNone         SymbolMode = ""
	SymbolModeRead         SymbolMode = "read"
	SymbolModeReadSelected SymbolMode = "readSelected"
	SymbolModeReadAll      SymbolMode = "readAll" // read all symbols
)

// Input is sythesis input data
type Input struct {
	Text string `json:"text,omitempty"`
	//TextType may have values: text, ssml
	TextType string `json:"textType,omitempty"`
	//Possible values are m4a, mp3, wav, ulaw
	OutputFormat     string  `json:"outputFormat,omitempty"`
	OutputTextFormat string  `json:"outputTextFormat,omitempty"`
	AllowCollectData *bool   `json:"saveRequest,omitempty"`
	Speed            float64 `json:"speed,omitempty"`
	Voice            string  `json:"voice,omitempty"`
	Priority         int     `json:"priority,omitempty"`
	//Possible values are: word
	SpeechMarkTypes      []string `json:"speechMarkTypes,omitempty"`
	MaxEdgeSilenceMillis *int64   `json:"maxEdgeSilenceMillis,omitempty"`

	SymbolMode      SymbolMode `json:"symbolMode,omitempty"`
	SelectedSymbols []string   `json:"selectedSymbols,omitempty"`
}

// SpeechMark
type SpeechMark struct {
	//In Millis from the start of the audio
	TimeInMillis int64 `json:"timeMillis" msgpack:"timeMillis,omitempty"`
	//In Millis
	Duration int64 `json:"durationMillis,omitempty" msgpack:"duration,omitempty"`
	//Possible values are: word
	Type  string `json:"type,omitempty" msgpack:"type,omitempty"`
	Value string `json:"value,omitempty" msgpack:"value,omitempty"`
}

// Result is synthesis result
type Result struct {
	AudioAsString string        `json:"audioAsString,omitempty" msgpack:"audioAsString,omitempty"`
	Audio         []byte        `json:"audio,omitempty" msgpack:"audio,omitempty"`
	Error         string        `json:"error,omitempty" msgpack:"error,omitempty"`
	Text          string        `json:"text,omitempty" msgpack:"text,omitempty"`
	RequestID     string        `json:"requestID,omitempty" msgpack:"requestID,omitempty"`
	SpeechMarks   []*SpeechMark `json:"speechMarks,omitempty" msgpack:"speechMarks,omitempty"`
}

// InfoResult is a response for /synthesizeInfo request
type InfoResult struct {
	Count int64 `json:"count"`
}
