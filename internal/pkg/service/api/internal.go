//go:generate stringer -type=AudioFormatEnum

package api

import (
	"strconv"

	"github.com/airenas/tts-line/pkg/ssml"
)

// TextFormatEnum represent possible output text types
type TextFormatEnum int

const (
	//TextNone do not output text
	TextNone TextFormatEnum = iota
	//TextNormalized output normalized text
	TextNormalized
	//TextAccented output normalized accented text
	TextAccented
	//TextTranscribed output data that was sent to AM
	TextTranscribed
)

func (e TextFormatEnum) String() string {
	if e < TextNone || e > TextAccented {
		return "TextFormatEnum:" + strconv.Itoa(int(e))
	}
	return [...]string{"", "normalized", "accented"}[e]
}

// AudioFormatEnum represent possible audio outputs
type AudioFormatEnum int

const (
	//AudioNone value
	AudioNone AudioFormatEnum = iota
	//AudioDefault value
	AudioDefault
	//AudioMP3 value
	AudioMP3
	//AudioM4A value`
	AudioM4A
	//AudioWAV value
	AudioWAV
	//AudioULAW value
	AudioULAW
)

// OutputContentTypeEnum represent possible service outputs
type OutputContentTypeEnum int

const (
	//ContentUnspecified value
	ContentUnspecified OutputContentTypeEnum = iota
	//ContentJSON value
	ContentJSON
	//ContentMsgPack value
	ContentMsgPack
)

// TTSRequestConfig config for request
type TTSRequestConfig struct {
	Text                 string
	RequestID            string
	OutputFormat         AudioFormatEnum
	OutputTextFormat     TextFormatEnum
	OutputContentType    OutputContentTypeEnum
	OutputMetadata       []string
	AllowCollectData     bool
	SaveTags             []string
	Speed                float32
	Voice                string
	Priority             int
	AllowedMaxLen        int
	SSMLParts            []ssml.Part
	AudioSuffix          string
	SpeechMarkTypes      map[string]bool
	MaxEdgeSilenceMillis int64
}
