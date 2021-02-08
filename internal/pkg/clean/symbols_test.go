package clean

import (
	"testing"

	"gotest.tools/assert"
)

func TestChangeSymbols(t *testing.T) {
	assert.Equal(t, "", changeSymbols(""))
	assert.Equal(t, "a a", changeSymbols("a a"))
	assert.Equal(t, "a - a -", changeSymbols("a – a –"))
	assert.Equal(t, "a--a", changeSymbols("a--a"))
	assert.Equal(t, "a\na", changeSymbols("a\na"))
	assert.Equal(t, "a\n\na\n\n", changeSymbols("a\n\ra\r\r"))
	assert.Equal(t, "a\n a", changeSymbols("a\n\ta"))
}

func TestChangeLetters(t *testing.T) {
	assert.Equal(t, "Janis", changeSymbols("Jānis"))
	assert.Equal(t, "Agrastas", changeSymbols("Ägrastas"))
	assert.Equal(t, "fizinės", changeSymbols("fizinės"))
}
