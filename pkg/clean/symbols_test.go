package clean

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/norm"
)

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
	// ts(t, "COVID", "СOVID")
	ts(t, "Erdoganas", "Erdoğanas")
	ts(t, "Zdenekas", "Zdeněkas")
	ts(t, "mokesčiai,", "mokesčiai֧")
	ts(t, "Kaune", "Кaune")
	ts(t, "acuza", "acuză")
	ts(t, "Vojtech", "Vojtěch")
	ts(t, "Eriktileri", "Erіktіlerі")
	ts(t, "Beiswenger", "Beiβwenger")
	ts(t, "a arba a", "a\u200carba\u200ca")
	ts(t, "Brulard", "Brûlard")
	ts(t, "saugios", "saugios֤")
	ts(t, "saugios\"", "saugios″")
	ts(t, "ATRAMOS", "𝗔𝗧𝗥𝗔𝗠𝗢𝗦")
}

func TestChangeSymbols(t *testing.T) {
	tests := []struct {
		name   string
		args   string
		want   string
		up, lw bool
	}{
		{name: "Empty", args: "", want: ""},
		{name: "a a", args: "a a", want: "a a", up: true, lw: true},
		{name: "a – a –", args: "a – a –", want: "a - a -", up: true, lw: true},
		{name: "a--a", args: "a--a", want: "a--a", up: true, lw: true},
		{name: "a——a", args: "a——a", want: "a--a", up: true, lw: true},
		{name: "a\na", args: "a\na", want: "a\na", up: true, lw: true},
		{name: "a\n\ra\r\r", args: "a\n\ra\r\r", want: "a\n\na\n\n", up: true, lw: true},
		{name: "a\n\ta", args: "a\n\ta", want: "a\n a", up: true, lw: true},

		{name: "Đuričić", args: "Đuričić", want: "Duričic", up: true, lw: true},
		{name: "pаgyvėjo", args: "pаgyvėjo", want: "pagyvėjo", up: true, lw: true},
		{name: "Košťal", args: "Košťal", want: "Koštal", up: true, lw: true},
		// {name: "В.vieną", args: "В.vieną", want: "B.vieną", up: true, lw: true},
		{name: "pi̇̀š", args: "pi̇̀š", want: "piš", up: true, lw: true},
		{name: "Moïsė", args: "Moïsė", want: "Moisė", up: true, lw: true},
		{args: "Tonyʼis", want: "Tonyis", up: true, lw: true},
		{args: "Tonyʼ is", want: "Tony' is", up: true, lw: true},
		{args: "aštuoniasdešimt′s", want: "aštuoniasdešimts", up: true, lw: true},
		{args: "aštuoniasdešimt′", want: "aštuoniasdešimt'", up: true, lw: true},
		{args: "Goethe´s", want: "Goethes", up: true, lw: true},
		{args: "Czesⱡawo", want: "Czeslawo", up: true, lw: true},
		{args: "Trenuleţul", want: "Trenuletul", up: true, lw: true},
		{args: "rugpjū́čio", want: "rugpjūčio", up: true, lw: true},
		{args: "į̃kvepia", want: "įkvepia", up: true, lw: true},
		{args: "kū́nas", want: "kūnas", up: true, lw: true},
		{args: "Κompozitoriai", want: "Kompozitoriai", up: true, lw: true},
		{args: "šeši∙dešimt", want: "šeši dešimt", up: true, lw: true},
		{args: "taŋka", want: "tanka", up: true, lw: true},
		{args: "Κartu", want: "Kartu", up: true, lw: true},
		{args: "Τiriami", want: "Tiriami", up: true, lw: true},
		{args: "Fazıl", want: "Fazil", up: true, lw: true},
		{args: "Εžeras", want: "Ežeras", up: true, lw: true},
		{args: "Talıs", want: "Talis", up: true, lw: true},
		{args: "Ηeterogeniniuose", want: "Heterogeniniuose", up: true, lw: true},
		{args: "Ιzoliacinė", want: "Izoliacinė", up: true, lw: true},
		{args: "SX→x→R", want: "SX-x-R", up: true, lw: true},
		{args: "kvœdi", want: "kvadi", up: true, lw: true},
		{args: "Αnsamblis", want: "Ansamblis", up: true, lw: true},
		{args: "Μelodiją", want: "Melodiją", up: true, lw: true},
		{args: "minutės→penkios", want: "minutės-penkios", up: true, lw: true},
		{args: "Тiškevičiai", want: "Tiškevičiai", up: true, lw: true},
		{args: "Οperacijų", want: "Operacijų", up: true, lw: true},
		{args: "trijų⋅tūkstančio", want: "trijų tūkstančio", up: true, lw: true},
		{args: "mutacĳos", want: "mutacijos", up: true, lw: true},
		{args: "dešimt⋅", want: "dešimt ", up: true, lw: true},
		{args: "Maître", want: "Maitre", up: true, lw: true},
		{args: "Feďkovyč", want: "Fedkovyč", up: true, lw: true},
		{args: "ꞌStaburadzeeꞌ", want: "'Staburadzee'", up: true, lw: true},
		{args: "septynis‰", want: "septynis ", up: true, lw: true},
		{args: "ruguǒ", want: "ruguo", up: true, lw: true},
		{args: "wǒmen", want: "women", up: true, lw: true},
		{args: "goumǎi", want: "goumai", up: true, lw: true},
		{args: "Ysaÿe", want: "Ysaye", up: true, lw: true},
		{args: "Νiekas", want: "Niekas", up: true, lw: false},
		// {args: "Нaрbe", want: "Hapbe", up: true, lw: false},
		{args: "Andruța", want: "Andruta", up: true, lw: true},
		{args: "Dîner", want: "Diner", up: true, lw: true},
		{args: "įtraukiájam", want: "įtraukiajam", up: true, lw: true},
		{args: "\u0002start", want: "start", up: true, lw: true},
		{args: "Ľubomiras", want: "Lubomiras", up: true, lw: true},
		{args: "tinklapisꓽ", want: "tinklapis:", up: true, lw: true},
		{args: "ÖVP ÄVP", want: "OVP AVP", up: true, lw: true},
		{args: "Kȩstutis", want: "Kęstutis", up: true, lw: true},
		{args: "ʽačiū", want: "'ačiū", up: true, lw: true},
		{args: "Дякую!", want: "Djakuju!", up: true, lw: true},
		{args: "Кириллица!", want: "Kirillica!", up: true, lw: true},
		{args: "H₂O", want: "H2O", up: true, lw: true},
		{args: "H₂O₃", want: "H2O3", up: true, lw: true},
	}
	for _, tt := range tests {
		n := tt.name
		if n == "" {
			n = tt.args
		}
		t.Run(n, func(t *testing.T) {
			testChange(t, tt.want, tt.args, tt.up, tt.lw)
		})
	}
}

func ts(t *testing.T, expected, inp string) {
	t.Helper()
	testChange(t, expected, inp, true, false)
}

func testChange(t *testing.T, expected, inp string, up, lw bool) {
	t.Helper()
	assert.Equal(t, expected, ChangeSymbols(inp))
	if up {
		assert.Equal(t, strings.ToUpper(expected), ChangeSymbols(strings.ToUpper(inp)))
	}
	if lw {
		assert.Equal(t, strings.ToLower(expected), ChangeSymbols(strings.ToLower(inp)))
	}

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
	str := "į̃kvepia"
	for i, r := range str {
		fmt.Printf("%d: %s %d \\u%.4x\n", i, string(r), r, r)
	}
	for i, r := range []rune(str) {
		fmt.Printf("%d: %s %d \\u%.4x\n", i, string(r), r, r)
	}
	sn := norm.NFC.String(strings.ToLower(str))
	fmt.Printf("str = %s (%d)\n", str, len(str))
	for _, r := range str {
		fmt.Printf("%s %d \\u%.4x\n", string(r), r, r)
	}
	fmt.Printf("str = %s\n", norm.NFC.String(strings.ToUpper(str)))
	for i, r := range norm.NFC.String(strings.ToUpper(str)) {
		fmt.Printf("%d: %s %d \\u%.4x\n", i, string(r), r, r)
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

func Test_dropOneQuote(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "Empty", args: "", want: ""},
		{name: "Quote", args: "Ąaewe'as", want: "Ąaeweas"},
		{name: "Several quotes", args: "Ąaewe'as\n aaaa'bbb", want: "Ąaeweas\n aaaabbb"},
		{name: "Lt", args: "Ą'č", want: "Ąč"},
		{name: "Leave", args: "`aaa", want: "`aaa"},
		{name: "Leave", args: "`aaa'", want: "`aaa'"},
		{name: "Leave", args: "aaa'", want: "aaa'"},
		{name: "Leave", args: " `aaa", want: " `aaa"},
		{name: "Leave", args: "aa' a", want: "aa' a"},
		{name: "Leave", args: "aa\"a", want: "aa\"a"},
		{name: "Leave", args: "aa`a", want: "aa`a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropOneQuote(tt.args); got != tt.want {
				t.Errorf("dropOneQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}
