package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert.Equal(t, "", normalize(""))
	assert.Equal(t, "a a", normalize("a a"))
	assert.Equal(t, "a - a -", normalize("        a     –       a     –   "))
	assert.Equal(t, "a--a", normalize("a--a"))
	assert.Equal(t, "a a", normalize("a\na"))
	assert.Equal(t, "a a", normalize("a\n\ra\r\r"))
	assert.Equal(t, "a a", normalize("a\n\ta"))
}

func TestNormalizeProcess(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	err := NewNormalizer().Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "a a", d.Text)
}

func TestNormalizeProcess_NoText(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.OriginalText = ""
	err := NewNormalizer().Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "no_text", d.ValidationFailures[0].Check.ID)
}

func TestNormalize_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	err := NewNormalizer().Process(d)
	assert.Nil(t, err)
}
