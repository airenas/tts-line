package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextFormatEnumString(t *testing.T) {
	assert.Equal(t, "TextFormatEnum:-1", TextFormatEnum(-1).String())
	assert.Equal(t, "", TextNone.String())
	assert.Equal(t, "normalized", TextNormalized.String())
	assert.Equal(t, "accented", TextAccented.String())
	assert.Equal(t, "TextFormatEnum:100", TextFormatEnum(100).String())
}

func TestAudioFormatEnumString(t *testing.T) {
	assert.Equal(t, "AudioFormatEnum(-1)", AudioFormatEnum(-1).String())
	assert.Equal(t, "AudioNone", AudioNone.String())
	assert.Equal(t, "AudioDefault", AudioDefault.String())
	assert.Equal(t, "AudioMP3", AudioMP3.String())
	assert.Equal(t, "AudioM4A", AudioM4A.String())
	assert.Equal(t, "AudioFormatEnum(100)", AudioFormatEnum(100).String())
}
