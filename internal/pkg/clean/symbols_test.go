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
	assert.Equal(t, "energiją", changeSymbols("energiją"))
	assert.Equal(t, "strategija-", changeSymbols("strategija―"))
	assert.Equal(t, "'Freda'", changeSymbols("ˈFredaˈ"))
	assert.Equal(t, "Francois", changeSymbols("François"))
	assert.Equal(t, "'Landė'", changeSymbols("ˈLandė'"))
	assert.Equal(t, "maculelė", changeSymbols("maculelê"))
	assert.Equal(t, "mūro", changeSymbols("mūro"))
	assert.Equal(t, "Garcia", changeSymbols("García"))
	assert.Equal(t, "Powrot", changeSymbols("Powrόt"))
	assert.Equal(t, "šešiasdešimt ", changeSymbols("šešiasdešimt˚"))
	assert.Equal(t, "Valstiečių", changeSymbols("Valstiečių"))
}
