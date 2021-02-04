package clean

var replaceableSymbols map[rune][]rune

func init() {
	replaceableSymbols = make(map[rune][]rune)
	for _, r := range []rune(" \t•\uFEFF\x00\u007f ·­") {
		replaceableSymbols[r] = []rune(" ")
	}
	replaceableSymbols['–'] = []rune("-")
	replaceableSymbols['…'] = []rune("...")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['‘'] = []rune("`")
	replaceableSymbols['”'] = []rune("\"")
	replaceableSymbols['‑'] = []rune("-")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['\r'] = []rune("\n")
	replaceableSymbols['\r'] = []rune("\n")
	for k, v := range getLettersMap() {
		replaceableSymbols[k] = []rune{v}
	}
}

func getLettersMap() map[rune]rune {
	res := make(map[rune]rune)
	res[193] = 'A'
	res[195] = 'A'
	res[224] = 'a'
	res[225] = 'a'
	res[227] = 'a'
	res[226] = 'a'
	res[236] = 'i'
	res[297] = 'i'
	res[204] = 'I'
	res[232] = 'e'
	res[242] = 'o'
	res[7929] = 'y'
	res[7869] = 'e'
	res[241] = 'n'
	res[249] = 'u'
	res[361] = 'u'
	res[237] = 'i'
	res[210] = 'O'
	res[205] = 'I'
	res[250] = 'u'
	res[253] = 'y'
	res['õ'] = 'o'
	res['ó'] = 'o'
	res['é'] = 'e'
	res['Õ'] = 'O'
	res['Å'] = 'A'
	res['Ä'] = 'A'
	res['Ã'] = 'A'
	res['Â'] = 'A'
	res['å'] = 'a'
	res['ã'] = 'a'
	res['â'] = 'a'
	res['ä'] = 'a'
	res['ā'] = 'a'
	return res
}

func changeSymbols(line string) string {
	if len(line) == 0 {
		return line
	}
	runes := []rune(line)
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
