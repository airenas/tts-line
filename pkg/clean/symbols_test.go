package clean

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/norm"
)

func TestChangeSymbols(t *testing.T) {
	assert.Equal(t, "", ChangeSymbols(""))
	assert.Equal(t, "a a", ChangeSymbols("a a"))
	assert.Equal(t, "a - a -", ChangeSymbols("a – a –"))
	assert.Equal(t, "a--a", ChangeSymbols("a--a"))
	assert.Equal(t, "a--a", ChangeSymbols("a——a"))
	assert.Equal(t, "a--a", ChangeSymbols("a--a"))
	assert.Equal(t, "a\na", ChangeSymbols("a\na"))
	assert.Equal(t, "a\n\na\n\n", ChangeSymbols("a\n\ra\r\r"))
	assert.Equal(t, "a\n a", ChangeSymbols("a\n\ta"))
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
	ts(t, "hi", "h\u200di")
	ts(t, "hi\"", "hi‟")
	ts(t, "inios", "˛inios")
	ts(t, "'Kanklės", "‚Kanklės")
	ts(t, "trijų/keturių", "trijų⁄keturių")
	ts(t, "aštuoni⁰", "aštuoni⁰")
	ts(t, "septyni⁰C", "septyni℃")
	assert.Equal(t, "fiksuota", ChangeSymbols("ﬁksuota"))
	ts(t, "Nasdaq\"", "Nasdaq¨")
	ts(t, "Hustiu", "Huștiu")
	ts(t, "Katerina", "Kateřina")
	ts(t, "Tony'is", "Tonyʼis")
	ts(t, "COVID", "СOVID")
	ts(t, "Erdoganas", "Erdoğanas")
	ts(t, "Zdenekas", "Zdeněkas")
	ts(t, "mokesčiai,", "mokesčiai֧")
	ts(t, "aštuoniasdešimt's", "aštuoniasdešimt′s")
	ts(t, "Kaune", "Кaune")
	ts(t, "acuza", "acuză")
	ts(t, "Vojtech", "Vojtěch")
	ts(t, "Eriktileri", "Erіktіlerі")
	ts(t, "Beiswenger", "Beiβwenger")
	ts(t, "a arba a", "a\u200carba\u200ca")
	ts(t, "Brulard", "Brûlard")
	ts(t, "saugios", "saugios֤")
	ts(t, "saugios\"", "saugios″")
	ts(t, "pagyvėjo", "pаgyvėjo")
	ts(t, "Duričić", "Đuričić")
}

func ts(t *testing.T, expected, inp string) {
	assert.Equal(t, expected, ChangeSymbols(inp))
	up := strings.ToUpper(inp)
	assert.Equal(t, strings.ToUpper(expected), ChangeSymbols(up))
}

func TestDash(t *testing.T) {
	for _, s := range []string{"-", "‒", "–"} {
		assert.Equal(t, "-", ChangeSymbols(s))
	}
}

func TestSymbols(t *testing.T) {
	str := ChangeSymbols("ą̨  a\u0328")
	for _, r := range str {
		fmt.Printf("a%s %d \\u%.4x\n", string(r), r, r)
	}
}

func TestSymbols2(t *testing.T) {
	str := "◌"
	sn := norm.NFC.String(strings.ToLower(str))
	fmt.Printf("str = %s (%d)\n", str, len(str))
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