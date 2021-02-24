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
	assert.Equal(t, "fų", changeSymbols("fų"))
	assert.Equal(t, "fš", changeSymbols("fš"))
	assert.Equal(t, "fc", changeSymbols("fcׅ"))
	assert.Equal(t, "faffa", changeSymbols("fa\u200b\u200b\u200bffa"))
	assert.Equal(t, "ojo", changeSymbols("оjо"))
}
