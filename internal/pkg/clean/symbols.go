package clean

import (
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var replaceableSymbols map[rune][]rune

func init() {
	replaceableSymbols = make(map[rune][]rune)
	for _, r := range []rune("\u200b¡\u05c5\u0328\u200d˛\u05a4") { // drop symbols
		replaceableSymbols[r] = []rune{}
	}
	replaceableSymbols['⎯'] = []rune("_")
	replaceableSymbols['…'] = []rune("...")
	replaceableSymbols['‘'] = []rune("`")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['℃'] = []rune("⁰C")
	replaceableSymbols['ﬁ'] = []rune("fi")

	for k, v := range getMaps() {
		replaceableSymbols[k] = []rune{v}
	}
}

func getMaps() map[rune]rune {
	res := make(map[rune]rune)

	addMap(res, "'ˈ‚ʼ′", '\'')
	addMap(res, "”‟¨″", '"')
	addMap(res, "\u2028\uFEFF\u001e\x00\u007f\t•·­˚\u200c", ' ')
	addMap(res, "–—―‐‑‒", '-')
	addMap(res, "⁄", '/')

	addMap(res, "С", 'C')
	addMap(res, "К", 'K')
	addMap(res, "\u05a7", ',')

	addLetterMap(res, "āäâãåàáăа", 'a')
	addLetterMap(res, "оôõόóőò", 'o')
	addLetterMap(res, "ç", 'c')
	addLetterMap(res, "еéëèẽəě", 'e')
	addLetterMap(res, "ê", 'ė')
	addLetterMap(res, "ð", 'd')
	addLetterMap(res, "ůùũúû", 'u')
	addLetterMap(res, "ìĩíі", 'i')
	addLetterMap(res, "ýỹ", 'y')
	addLetterMap(res, "ñ", 'n')
	addLetterMap(res, "ř", 'r')
	addLetterMap(res, "şșβ", 's')
	addLetterMap(res, "ğ", 'g')
	addLetterMap(res, "đ", 'd')

	return res
}

func addLetterMap(res map[rune]rune, add string, to rune) {
	for _, r := range add {
		res[r] = to
		res[unicode.ToUpper(r)] = unicode.ToUpper(to)
	}
}

func addMap(res map[rune]rune, add string, to rune) {
	for _, r := range add {
		res[r] = to
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
