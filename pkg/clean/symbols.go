package clean

import (
	"regexp"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var replaceableSymbols map[rune][]rune
var oneQuoteBetweenLetters *regexp.Regexp

func init() {
	replaceableSymbols = make(map[rune][]rune)
	for _, r := range "\u200b¡\u05c5\u0328\u200d˛\u05a4\u0307\u0300\u0301\u0303" { // drop symbols
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
	oneQuoteBetweenLetters = regexp.MustCompile(`(\p{L})'(\p{L})`)
}

func getMaps() map[rune]rune {
	res := make(map[rune]rune)

	addMap(res, "'ˈ‚ʼ′´", '\'')
	addMap(res, "”‟¨″", '"')
	addMap(res, "\u2028\uFEFF\u001e\x00\u007f\t•·­˚\u200c", ' ')
	addMap(res, "–—―‐‑‒", '-')
	addMap(res, "⁄", '/')

	addMap(res, "С", 'C')
	addMap(res, "К", 'K')
	addMap(res, "\u05a7", ',')
	addMap(res, "\u0130", 'I')

	addLetterMap(res, "āäâãåàáăа", 'a')
	addLetterMap(res, "оôõόóőò", 'o')
	addLetterMap(res, "çć", 'c')
	addLetterMap(res, "еéëèẽəě", 'e')
	addLetterMap(res, "ê", 'ė')
	addLetterMap(res, "ð", 'd')
	addLetterMap(res, "ůùũúû", 'u')
	addLetterMap(res, "ìĩíіï", 'i')
	addLetterMap(res, "ýỹ", 'y')
	addLetterMap(res, "ñ", 'n')
	addLetterMap(res, "ř", 'r')
	addLetterMap(res, "şșβ", 's')
	addLetterMap(res, "ğ", 'g')
	addLetterMap(res, "đ", 'd')
	addLetterMap(res, "ťţ", 't')
	addLetterMap(res, "в", 'b')
	addLetterMap(res, "ⱡ", 'l')
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

// ChangeSymbols updates symbols to be compatible with standard Lithuanian alphabet
// Also some rare quotes are unified to the most popular ones
func ChangeSymbols(line string) string {
	if len(line) == 0 {
		return line
	}
	lineU := norm.NFC.String(line)
	res := make([]rune, 0)
	for _, r := range lineU {
		res = append(res, changeSymbol(r)...)
	}
	return dropOneQuote(string(res))
}

func changeSymbol(r rune) []rune {
	s, ok := replaceableSymbols[r]
	if ok {
		return s
	}
	return []rune{r}
}

func dropOneQuote(in string) string {
	return oneQuoteBetweenLetters.ReplaceAllString(in, "$1$2")
}
