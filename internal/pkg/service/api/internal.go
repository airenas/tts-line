package api

import "strconv"

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

func (e TextFormatEnum) String() string {
	if e < TextNone || e > TextAccented {
		return "TextFormatEnum:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "normalized", "accented"}[e]
}

//AudioFormatEnum represent possible audio outputs
type AudioFormatEnum int

const (
	//AudioNone value
	AudioNone AudioFormatEnum = iota
	//AudioMP3 value
	AudioMP3
	//AudioM4A value`
	AudioM4A
)

func (e AudioFormatEnum) String() string {
	if e < AudioNone || e > AudioM4A {
		return "AudioFormatEnum:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "mp3", "m4a"}[e]
}

//TTSRequestConfig config for request`
type TTSRequestConfig struct {
	Text             string
	OutputFormat     AudioFormatEnum
	OutputTextFormat TextFormatEnum
	OutputMetadata   []string
	AllowCollectData bool
}
