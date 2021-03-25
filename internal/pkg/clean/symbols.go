package clean

import (
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var replaceableSymbols map[rune][]rune

func init() {
	replaceableSymbols = make(map[rune][]rune)
	for _, r := range []rune("\u2028\uFEFF\u001e\x00\u007f\t•·­˚") {
		replaceableSymbols[r] = []rune(" ")
	}
	for _, r := range []rune("–—―‐‑‒") {
		replaceableSymbols[r] = []rune("-")
	}
	for _, r := range []rune("\u200b¡\u05c5\u0328") { // drop symbols
		replaceableSymbols[r] = []rune{}
	}
	replaceableSymbols['⎯'] = []rune("_")
	replaceableSymbols['…'] = []rune("...")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['‘'] = []rune("`")
	replaceableSymbols['”'] = []rune("\"")
	replaceableSymbols['\r'] = []rune("\n")

	for k, v := range getLettersMap() {
		replaceableSymbols[k] = []rune{v}
	}
}

func getLettersMap() map[rune]rune {
	res := make(map[rune]rune)
	res['\''] = '\''
	res['ˈ'] = '\''
	addLetterMap(res, "āäâãåàá", 'a')
	addLetterMap(res, "оôõόóőò", 'o')
	addLetterMap(res, "ç", 'c')
	addLetterMap(res, "еéëèẽ", 'e')
	addLetterMap(res, "ê", 'ė')
	addLetterMap(res, "ð", 'd')
	addLetterMap(res, "ůùũú", 'u')
	addLetterMap(res, "ìĩí", 'i')
	addLetterMap(res, "ýỹ", 'y')
	addLetterMap(res, "ñ", 'n')

	return res
}

func addLetterMap(res map[rune]rune, add string, to rune) {
	for _, r := range add {
		res[r] = to
	}
	for _, r := range add {
		res[unicode.ToUpper(r)] = unicode.ToUpper(to)
	}
}

func changeSymbols(line string) string {
	if len(line) == 0 {
		return line
	}
	lineU := norm.NFC.String(line)
	runes := []rune(lineU)
	res := make([]rune, 0)
	for _, r := range runes {
		res = append(res, changeSymbol(r)...)
	}
	return string(res)
}

func changeSymbol(r rune) []rune {
	s, ok := replaceableSymbols[r]
	if ok {
		return s
	}
	return []rune{r}
}
