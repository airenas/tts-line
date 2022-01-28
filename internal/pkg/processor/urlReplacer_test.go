package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReplacer(t *testing.T) {
	pr := NewURLReplacer().(*urlReplacer)
	require.NotNil(t, pr)
	assert.Equal(t, "Internetinis adresas", pr.phrase)
}

func TestReplacerProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.CleanedText = " a a www.delfi.lt"
	pr := NewURLReplacer()
	require.NotNil(t, pr)
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, " a a Internetinis adresas", d.Text)
}

func TestReplacer_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	d.CleanedText = "text"
	pr := NewURLReplacer()
	err := pr.Process(d)
	require.Nil(t, err)
	assert.Equal(t, "", d.Text)
}
