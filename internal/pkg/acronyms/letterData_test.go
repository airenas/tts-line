package acronyms

import (
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
)

func Test_readVocab(t *testing.T) {
	tests := []struct {
		name   string
		symbol string
		wanted ldata
		modes  []api.Mode
	}{
		{name: "comma", symbol: ",", wanted: ldata{ch: "kablelis", letter: "kablelis", newWord: true}, modes: []api.Mode{api.ModeAllAsCharacters}},
		{name: "1", symbol: "1", wanted: ldata{ch: "vienas", letter: "vienas", newWord: true}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
		{name: "symbol", symbol: "$", wanted: ldata{ch: "doleris", letter: "doleris", newWord: true}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
		{name: "lower", symbol: "α", wanted: ldata{ch: "alfa", letter: "alfa", newWord: true}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
		{name: "upper", symbol: "Α", wanted: ldata{ch: "alfa", letter: "alfa", newWord: true}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
		{name: "several", symbol: "®", wanted: ldata{ch: "registruotas", letter: "registruotas", newWord: true,
			next: &ldata{ch: "prekės", letter: "prekės", newWord: true,
				next: &ldata{ch: "ženklas", letter: "ženklas", newWord: true}}}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
		{name: "several", symbol: "™", wanted: ldata{ch: "prekės", letter: "prekės", newWord: true,
			next: &ldata{ch: "ženklas", letter: "ženklas", newWord: true}}, modes: []api.Mode{api.ModeAllAsCharacters, api.ModeCharacters}},
	}
	res := newLettersMap()
	gotErr := readVocab(pronunciationCSV, &res)
	if gotErr != nil {
		t.Errorf("readVocab() failed: %v", gotErr)
		return
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, mode := range [...]api.Mode{api.ModeCharacters, api.ModeCharactersAsWord, api.ModeAllAsCharacters} {
				got, ok := res[mode][tt.symbol]
				if !testHasMode(tt.modes, mode) {
					if ok {
						t.Errorf("readVocab() unexpected data for mode %v and symbol %v", mode, tt.symbol)
					}
					continue
				}
				if !ok {
					t.Errorf("readVocab() no data for mode %v and symbol %v", mode, tt.symbol)
					continue
				}
				testLdataCmp(t, got, &tt.wanted)
			}
		})
	}
}

func testLdataCmp(t *testing.T, got *ldata, wanted *ldata) {
	if got == nil && wanted == nil {
		return
	}
	if got == nil {
		t.Errorf("readVocab() = %v, want %v", got, *wanted)
		return
	}
	if wanted == nil {
		t.Errorf("readVocab() = %v, want %v", *got, wanted)
		return
	}
	g := *got
	w := *wanted
	g.next = nil
	w.next = nil
	if !reflect.DeepEqual(g, w) {
		t.Errorf("readVocab() = %v, want %v", g, w)
	}
	testLdataCmp(t, got.next, wanted.next)
}

func testHasMode(mode1 []api.Mode, mode2 api.Mode) bool {
	for _, m := range mode1 {
		if m == mode2 {
			return true
		}
	}
	return false
}
