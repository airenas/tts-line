package clean

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
	"gotest.tools/assert"
)

func TestChangeSymbols(t *testing.T) {
	assert.Equal(t, "", changeSymbols(""))
	assert.Equal(t, "a a", changeSymbols("a a"))
	assert.Equal(t, "a - a -", changeSymbols("a – a –"))
	assert.Equal(t, "a--a", changeSymbols("a--a"))
	assert.Equal(t, "a--a", changeSymbols("a——a"))
	assert.Equal(t, "a--a", changeSymbols("a--a"))
	assert.Equal(t, "a\na", changeSymbols("a\na"))
	assert.Equal(t, "a\n\na\n\n", changeSymbols("a\n\ra\r\r"))
	assert.Equal(t, "a\n a", changeSymbols("a\n\ta"))
}

func TestChangeLetters(t *testing.T) {
	ts(t, "Janis", "Jānis")
	ts(t, "Agrastas", "Ägrastas")
	ts(t, "fizinės", "fizinės")
	ts(t, "fų", "fų")
	ts(t, "fš", "fš")
	ts(t, "fc", "fcׅ")
	ts(t, "faffa", "fa\u200b\u200b\u200bffa")
	ts(t, "ojo", "оjо")
	ts(t, "energiją", "energiją")
	ts(t, "strategija-", "strategija―")
	ts(t, "'Freda'", "ˈFredaˈ")
	ts(t, "Francois", "François")
	ts(t, "'Landė'", "ˈLandė'")
	ts(t, "maculelė", "maculelê")
	ts(t, "mūro", "mūro")
	ts(t, "Garcia", "García")
	ts(t, "Powrot", "Powrόt")
	ts(t, "šešiasdešimt ", "šešiasdešimt˚")
	ts(t, "Valstiečių", "Valstiečių")
	ts(t, "įstatymo", "įstatymo")
	ts(t, "grįžo", "grįžo")
	ts(t, "pratęsė", "pratęsė")
	ts(t, "Gagnamagnid", "Gagnamagnið")
	ts(t, "Hruša", "Hrůša")
	ts(t, "Tudos", "Tudős")
	ts(t, "Citroen", "Citroën")
	ts(t, "stresą", "stresą̨")
	ts(t, "paviršius", "paviršiùs")
	ts(t, "semeni", "səməni")
	ts(t, "hsi", "hşi")
}

func ts(t *testing.T, expected, inp string) {
	assert.Equal(t, expected, changeSymbols(inp))
	up := strings.ToUpper(inp)
	assert.Equal(t, strings.ToUpper(expected), changeSymbols(up))
}

func TestDash(t *testing.T) {
	for _, s := range []string{"-", "‒", "–"} {
		assert.Equal(t, "-", changeSymbols(s))
	}
}

func TestSymbols(t *testing.T) {
	str := changeSymbols("ą̨  a\u0328")
	for _, r := range str {
		fmt.Printf("a%s %d \\u%.4x\n", string(r), r, r)
	}
}

func TestSymbols2(t *testing.T) {
	str := "\u00f9ū" + string(rune(241)) + string(rune(7929))
	sn := norm.NFC.String(str)
	fmt.Printf("str = %s\n", str)
	for _, r := range str {
		fmt.Printf("%s %d \\u%.4x\n", string(r), r, r)
	}
	fmt.Printf("str = %s\n", strings.ToUpper(str))
	for _, r := range strings.ToUpper(str) {
		fmt.Printf("%s %d \\u%.4x\n", string(r), r, r)
	}

	fmt.Printf("sn  = %s\n", sn)
	for _, r := range sn {
		fmt.Printf("%s %d \\u%.4x\n", string(r), r, r)
	}
}
