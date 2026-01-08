package acronyms

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/transcription"
)

// var letters map[string]*ldata

type ldata struct {
	ch       string
	chAccent string
	letter   string
	newWord  bool
	next     *ldata
}

type charType int

const (
	charTypeRegular charType = iota
	charTypeLong
	charTypeDiacritic
)

type pronData struct {
	char, pron string
	ct         charType
}

func newLettersMap() map[api.Mode]map[string]*ldata {
	res := make(map[api.Mode]map[string]*ldata)
	res[api.ModeCharacters] = map[string]*ldata{}
	res[api.ModeCharactersAsWord] = map[string]*ldata{}
	res[api.ModeAllAsCharacters] = map[string]*ldata{}
	return res
}

func initLetters() map[api.Mode]map[string]*ldata {
	res := newLettersMap()

	oneChar := letterProns()
	for _, c := range oneChar {
		add(&res, c.char, makeLD(c.pron), api.ModeCharactersAsWord)
		add(&res, c.char, wordFromPron(c), api.ModeCharacters, api.ModeAllAsCharacters)
	}

	add(&res, "ch", makeLD("chą3"), api.ModeCharactersAsWord)
	add(&res, "dz", makeLD("dzė3"), api.ModeCharactersAsWord)
	add(&res, "dž", makeLD("džė3"), api.ModeCharactersAsWord)

	// w
	w := word("da4bl-vė")
	w.letter = "w"
	add(&res, "w", w, api.ModeCharactersAsWord, api.ModeCharacters, api.ModeAllAsCharacters)

	add(&res, ".", word("ta3-škas"), api.ModeCharactersAsWord, api.ModeAllAsCharacters)

	if err := readVocab(pronunciationCSV, &res); err != nil {
		log.Fatal().Err(err).Msg("read pronunciation data")
	}

	return res
}

//go:embed data/pronunciation.csv
var pronunciationCSV string

func readVocab(data string, res *map[api.Mode]map[string]*ldata) error {
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	for i, rec := range records {
		if i == 0 {
			continue // header
		}
		s := rec[0]
		if s == "" {
			continue
		}
		tp := strings.TrimSpace(rec[1])
		base, err := parseBool(rec[3])
		if err != nil {
			return fmt.Errorf("invalid base flag '%s' in line %d: %v", rec[3], i+1, err)
		}
		isLetter, err := parseBool(rec[4])
		if err != nil {
			return fmt.Errorf("invalid isLetter flag '%s' in line %d: %v", rec[4], i+1, err)
		}
		var ld *ldata
		switch tp {
		case "f": // full
			ld = words(strings.Split(rec[2], " ")...)
		case "":
			ld = wordsSimple(strings.Split(rec[2], " ")...)
		default:
			return fmt.Errorf("unknown type '%s' in line %d", tp, i+1)
		}
		mode := []api.Mode{api.ModeAllAsCharacters}
		if !base {
			mode = append(mode, api.ModeCharacters)
		}
		if isLetter {
			add(res, strings.ToLower(s), ld, mode...)
			add(res, strings.ToUpper(s), ld, mode...)
		} else {
			add(res, s, ld, mode...)
		}
	}
	return nil
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(s)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}

func letterProns() []pronData {
	var res []pronData
	res = append(res, pronData{char: "a", pron: "ą3"})
	res = append(res, pronData{char: "ą", pron: "ą3", ct: charTypeDiacritic})
	res = append(res, pronData{char: "b", pron: "bė3"})
	res = append(res, pronData{char: "c", pron: "cė3"})
	res = append(res, pronData{char: "č", pron: "čė3"})
	res = append(res, pronData{char: "d", pron: "dė3"})
	res = append(res, pronData{char: "e", pron: "ę3"})
	res = append(res, pronData{char: "ę", pron: "ę3", ct: charTypeDiacritic})
	res = append(res, pronData{char: "ė", pron: "ė3"})
	res = append(res, pronData{char: "f", pron: "e4f"})
	res = append(res, pronData{char: "g", pron: "gė3"})
	res = append(res, pronData{char: "h", pron: "hą3"})
	res = append(res, pronData{char: "i", pron: "į3"})
	res = append(res, pronData{char: "į", pron: "į3", ct: charTypeDiacritic})
	res = append(res, pronData{char: "y", pron: "į3", ct: charTypeLong})
	res = append(res, pronData{char: "j", pron: "jO4t"})
	res = append(res, pronData{char: "k", pron: "ką3"})
	res = append(res, pronData{char: "l", pron: "el3"})
	res = append(res, pronData{char: "m", pron: "em3"})
	res = append(res, pronData{char: "n", pron: "en3"})
	res = append(res, pronData{char: "o", pron: "o3"})
	res = append(res, pronData{char: "p", pron: "pė3"})
	res = append(res, pronData{char: "q", pron: "kų3"})
	res = append(res, pronData{char: "r", pron: "er3"})
	res = append(res, pronData{char: "s", pron: "e4s"})
	res = append(res, pronData{char: "š", pron: "e4š"})
	res = append(res, pronData{char: "t", pron: "tė3"})
	res = append(res, pronData{char: "u", pron: "ų3"})
	res = append(res, pronData{char: "ū", pron: "ų3", ct: charTypeLong})
	res = append(res, pronData{char: "ų", pron: "ų3", ct: charTypeDiacritic})
	res = append(res, pronData{char: "v", pron: "vė3"})
	res = append(res, pronData{char: "x", pron: "i4ks"})
	res = append(res, pronData{char: "z", pron: "zė3"})
	res = append(res, pronData{char: "ž", pron: "žė3"})
	// // ch, dz, dž
	// res = append(res, pronData{char: "ch", pron: "chą3"})
	// res = append(res, pronData{char: "dz", pron: "dzė3"})
	// res = append(res, pronData{char: "dž", pron: "džė3"})
	return res
}

func add(res *map[api.Mode]map[string]*ldata, char string, ldata *ldata, mode ...api.Mode) {
	if ldata.letter == "" {
		ldata.letter = char
	} // else perhaps word
	for _, m := range mode {
		(*res)[m][char] = ldata
	}
}

func makeLD(ch string) *ldata {
	var r ldata
	r.chAccent = ch
	r.ch = transcription.TrimAccent(ch)
	return &r
}

func word(pron string) *ldata {
	res := makeLD(pron)
	res.newWord = true
	res.letter = strings.ReplaceAll(pron, "-", "")
	res.letter = transcription.TrimAccent(strings.ToLower(res.letter))
	return res
}

func wordFromPron(c pronData) *ldata {
	res := makeLD(c.pron)
	res.newWord = true
	switch c.ct {
	case charTypeDiacritic:
		res.next = word("no9-si-nė")
	case charTypeLong:
		res.next = word("il-go9-ji")
	}
	return res
}

func words(w ...string) *ldata {
	if len(w) == 0 {
		log.Fatal().Msg("no words") // allow as it is initialization error
	}
	res := word(w[0])
	add := res
	for _, wn := range w[1:] {
		nw := word(wn)
		add.next = nw
		add = nw
	}
	return res
}

func wordsSimple(w ...string) *ldata {
	if len(w) == 0 {
		log.Fatal().Msg("no words") // allow as it is initialization error
	}
	res := &ldata{ch: w[0]}
	res.newWord = true
	res.letter = w[0]
	add := res
	for _, wn := range w[1:] {
		nw := &ldata{ch: wn}
		nw.newWord = true
		nw.letter = wn
		add.next = nw
		add = nw
	}
	return res
}
