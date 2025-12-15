package processor

import (
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func Test_readSymbols_Process(t *testing.T) {
	tests := []struct {
		name        string
		data        *synthesizer.TTSData
		mode        api.SymbolMode
		wantedWords []synthesizer.TaggedWord
	}{
		{name: "empty skip", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: true}}, wantedWords: nil},
		{name: "skip", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: true,
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeRead}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedWords: []synthesizer.TaggedWord{{Word: "A"}, {Separator: ","}, {Separator: "("}, {Word: "olia"}}},
		{name: "skip mode none", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeNone}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedWords: []synthesizer.TaggedWord{{Word: "A"}, {Separator: ","}, {Separator: "("}, {Word: "olia"}}},
		{name: "change (", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeRead}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedWords: []synthesizer.TaggedWord{{Word: "A"}, {Separator: ","},
			{Word: "skliaustai"}, {Word: "atsidaro"},
			{Word: "olia"}}},
		{name: "change comma", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeReadSelected, SelectedSymbols: []string{","}}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedWords: []synthesizer.TaggedWord{{Word: "A"},
			{Word: "kablelis"},
			{Separator: "("},
			{Word: "olia"}}},
		{name: "change several", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeReadSelected, SelectedSymbols: []string{",", "("}}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedWords: []synthesizer.TaggedWord{{Word: "A"},
			{Word: "kablelis"},
			{Word: "skliaustai"}, {Word: "atsidaro"},
			{Word: "olia"}}},
	}
	p, _ := NewReadSymbols()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := p.Process(t.Context(), tt.data)
			if gotErr != nil {
				t.Errorf("Process() failed: %v", gotErr)
				return
			}
			testTagged(t, tt.data.Words, tt.wantedWords)
		})
	}
}

func testTagged(t *testing.T, processedWord []*synthesizer.ProcessedWord, taggedWord []synthesizer.TaggedWord) {
	t.Helper()

	if len(processedWord) != len(taggedWord) {
		t.Errorf("Length mismatch got %d, wanted %d", len(processedWord), len(taggedWord))
		return
	}
	for i := range processedWord {
		if !reflect.DeepEqual(processedWord[i].Tagged, taggedWord[i]) {
			t.Errorf("Word %d mismatch got %+v, wanted %+v", i, processedWord[i].Tagged, taggedWord[i])
		}
	}
}
