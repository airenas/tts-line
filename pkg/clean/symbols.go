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
	for _, r := range "\u200b¡\u05c5\u0328\u200d˛\u05a4\u0307\u0300\u0301\u0303\u0002" { // drop symbols
		replaceableSymbols[r] = []rune{}
	}
	replaceableSymbols['⎯'] = []rune("_")
	replaceableSymbols['…'] = []rune("...")
	// replaceableSymbols['‘'] = []rune("`")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['℃'] = []rune("⁰C")
	replaceableSymbols['ﬁ'] = []rune("fi")
	replaceableSymbols['ĳ'] = []rune("ij")
	replaceableSymbols['Ĳ'] = []rune("IJ")

	for k, v := range getMaps() {
		replaceableSymbols[k] = []rune{v}
	}
	oneQuoteBetweenLetters = regexp.MustCompile(`(\p{L})'(\p{L})`)

	addCyrillicSymbols(replaceableSymbols)
}

func getMaps() map[rune]rune {
	res := make(map[rune]rune)

	addMap(res, "'ˈ‚ʼ′´ꞌꞋ`ʽ‘", '\'')
	addMap(res, "”‟¨″", '"')
	addMap(res, "\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u202f\u205f\u3000", ' ')
	addMap(res, "\u2028\uFEFF\u001e\x00\u007f\t•·­˚\u200c∙⋅‰", ' ')
	addMap(res, "‐‑‒", '-')
	addMap(res, "⁄", '/')
	addMap(res, "ꓽ", ':')

	addMap(res, "С", 'C')
	addMap(res, "Н", 'H')
	addMap(res, "К", 'K')
	addMap(res, "\u05a7", ',')
	addMap(res, "\u0130", 'I')
	addMap(res, "𝗔", 'A')
	addMap(res, "𝗧", 'T')
	addMap(res, "𝗥", 'R')
	addMap(res, "𝗠", 'M')
	addMap(res, "𝗢", 'O')
	addMap(res, "𝗦", 'S')

	// addMap(res, "₀", '0')
	// addMap(res, "₁", '1')
	// addMap(res, "₂", '2')
	// addMap(res, "₃", '3')
	// addMap(res, "₄", '4')
	// addMap(res, "₅", '5')
	// addMap(res, "₆", '6')
	// addMap(res, "₇", '7')
	// addMap(res, "₈", '8')
	// addMap(res, "₉", '9')

	addLetterMap(res, "āäâãåàáăаœǎ", 'a')
	addLetterMap(res, "в", 'b')
	addLetterMap(res, "çć", 'c')
	addLetterMap(res, "ðđď", 'd')
	addLetterMap(res, "еéëèẽəě", 'e')
	addLetterMap(res, "ȩ", 'ę')
	addLetterMap(res, "ê", 'ė')
	addLetterMap(res, "ğ", 'g')
	addLetterMap(res, "ìĩíіïıî", 'i')
	addLetterMap(res, "ýỹÿ", 'y')
	addLetterMap(res, "ľⱡ", 'l')
	addLetterMap(res, "ñ", 'n')
	addLetterMap(res, "оôõóőòǒö", 'o')
	addLetterMap(res, "ό", 'ο') // Greek omicron with tonos - drop tone
	addLetterMap(res, "р", 'p')
	addLetterMap(res, "ř", 'r')
	addLetterMap(res, "şș", 's')
	addLetterMap(res, "ťţтț", 't')
	addLetterMap(res, "ůùũúû", 'u')

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

func addCyrillicSymbols(replaceableSymbols map[rune][]rune) {
	addLettersMap(replaceableSymbols, "ъ", "")
	addLettersMap(replaceableSymbols, "ь", "")
	addLettersMap(replaceableSymbols, "б", "b")
	addLettersMap(replaceableSymbols, "в", "v")
	addLettersMap(replaceableSymbols, "г", "g")
	addLettersMap(replaceableSymbols, "д", "d")
	addLettersMap(replaceableSymbols, "ж", "ž")
	addLettersMap(replaceableSymbols, "з", "z")
	addLettersMap(replaceableSymbols, "к", "k")
	addLettersMap(replaceableSymbols, "л", "l")
	addLettersMap(replaceableSymbols, "м", "m")
	addLettersMap(replaceableSymbols, "н", "n")
	addLettersMap(replaceableSymbols, "п", "p")
	addLettersMap(replaceableSymbols, "р", "r")
	addLettersMap(replaceableSymbols, "с", "s")
	addLettersMap(replaceableSymbols, "т", "t")
	addLettersMap(replaceableSymbols, "ф", "f")
	addLettersMap(replaceableSymbols, "х", "xh")
	addLettersMap(replaceableSymbols, "ц", "c")
	addLettersMap(replaceableSymbols, "ч", "č")
	addLettersMap(replaceableSymbols, "ш", "š")
	addLettersMap(replaceableSymbols, "щ", "šč")
	addLettersMap(replaceableSymbols, "а", "a")
	addLettersMap(replaceableSymbols, "е", "ie")
	addLettersMap(replaceableSymbols, "ё", "jo")
	addLettersMap(replaceableSymbols, "и", "i")
	addLettersMap(replaceableSymbols, "о", "o")
	addLettersMap(replaceableSymbols, "у", "u")
	addLettersMap(replaceableSymbols, "ы", "y")
	addLettersMap(replaceableSymbols, "э", "e")
	addLettersMap(replaceableSymbols, "ю", "ju")
	addLettersMap(replaceableSymbols, "я", "ja")
	addLettersMap(replaceableSymbols, "й", "j")
	addLettersMap(replaceableSymbols, "ґ", "g")
	addLettersMap(replaceableSymbols, "є", "je")
	addLettersMap(replaceableSymbols, "ї", "ji")
}

func addLettersMap(res map[rune][]rune, add string, to string) {
	for _, r := range add {
		rns := []rune(to)
		res[r] = rns
		res[unicode.ToUpper(r)] = capitalize(rns)
	}
}

func capitalize(rns []rune) []rune {
	if len(rns) == 0 {
		return []rune{}
	}
	res := make([]rune, len(rns))
	for i := 0; i < len(rns); i++ {
		res[i] = unicode.ToUpper(rns[i])
	}
	return res
}
