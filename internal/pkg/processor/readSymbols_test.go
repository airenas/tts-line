package processor

import (
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func Test_readSymbols_Process(t *testing.T) {
	tests := []struct {
		name      string
		data      *synthesizer.TTSData
		mode      api.SymbolMode
		wantedNER []synthesizer.NEREnum
	}{
		{name: "empty skip", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: true}}, wantedNER: nil},
		{name: "skip", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: true,
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeRead}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedNER: []synthesizer.NEREnum{
			synthesizer.NERRegular,
			synthesizer.NERRegular,
			synthesizer.NERRegular,
			synthesizer.NERRegular,
		}},
		{name: "skip mode none", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeNone}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedNER: []synthesizer.NEREnum{
			synthesizer.NERRegular,
			synthesizer.NERRegular,
			synthesizer.NERRegular,
			synthesizer.NERRegular,
		}},
		{name: "change (", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeRead}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedNER: []synthesizer.NEREnum{
			synthesizer.NERRegular,
			synthesizer.NERReadableSymbol,
			synthesizer.NERReadableSymbol,
			synthesizer.NERRegular,
		}},
		{name: "change comma", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeReadSelected, SelectedSymbols: []string{","}}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedNER: []synthesizer.NEREnum{
			synthesizer.NERRegular,
			synthesizer.NERReadableAllSymbol,
			synthesizer.NERRegular,
			synthesizer.NERRegular,
		}},
		{name: "change several", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{
			Input: &api.TTSRequestConfig{SymbolMode: api.SymbolModeReadSelected, SelectedSymbols: []string{",", "("}}},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Separator: "("}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			},
		}, wantedNER: []synthesizer.NEREnum{
			synthesizer.NERRegular,
			synthesizer.NERReadableAllSymbol,
			synthesizer.NERReadableAllSymbol,
			synthesizer.NERRegular,
		}},
	}
	p, _ := NewReadSymbols()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := p.Process(t.Context(), tt.data)
			if gotErr != nil {
				t.Errorf("Process() failed: %v", gotErr)
				return
			}
			testTagged(t, tt.data.Words, tt.wantedNER)
		})
	}
}

func testTagged(t *testing.T, processedWord []*synthesizer.ProcessedWord, ners []synthesizer.NEREnum) {
	t.Helper()

	if len(processedWord) != len(ners) {
		t.Errorf("Length mismatch got %d, wanted %d", len(processedWord), len(ners))
		return
	}
	for i := range processedWord {
		if !reflect.DeepEqual(processedWord[i].NERType, ners[i]) {
			t.Errorf("Word %d mismatch got %+v, wanted %+v", i, processedWord[i].NERType, ners[i])
		}
	}
}
