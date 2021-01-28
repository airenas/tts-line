package api

//TextFormatEnum represent possible output text types
type TextFormatEnum int

const (
	//TextNone do not output text
	TextNone TextFormatEnum = iota
	//TextNormalized output normalized text
	TextNormalized
	//TextAccented output normalized accented text
	TextAccented
)

//TTSRequestConfig config for request
type TTSRequestConfig struct {
	Text             string
	OutputFormat     string
	OutputTextFormat TextFormatEnum
	OutputMetadata   []string
	TextFormat       bool
	AllowCollectData bool
}
