package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func Test_processNERSentence(t *testing.T) {
	tests := []struct {
		name            string
		s               []*synthesizer.ProcessedWord
		wantedPosMarked []int
	}{
		{name: "nil", s: nil, wantedPosMarked: nil},
		{name: "empty", s: []*synthesizer.ProcessedWord{}, wantedPosMarked: []int{}},
		{name: "no change", s: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "aa"}},
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "aa"}}},
			wantedPosMarked: []int{}},
		{name: "Change", s: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "aa"}},
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "A"}}},
			wantedPosMarked: []int{2}},
		{name: "Change non first", s: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "A"}},
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "Ž"}}},
			wantedPosMarked: []int{2}},
		{name: "Non letter", s: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "A"}},
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "7"}}},
			wantedPosMarked: []int{}},
		{name: "Keep first", s: []*synthesizer.ProcessedWord{
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "Ž"}}},
			wantedPosMarked: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := processNERSentence(t.Context(), tt.s)
			if gotErr != nil {
				t.Errorf("processNERSentence() failed: %v", gotErr)
				return
			}
			testValidateNER(t, tt.s, tt.wantedPosMarked)
		})
	}
}

func Test_ner_Process(t *testing.T) {
	tests := []struct {
		name            string
		data            *synthesizer.TTSData
		wantedPosMarked []int
	}{
		{name: "skip", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: true}}, wantedPosMarked: nil},
		{name: "change", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: false},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{Word: "B"}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			}}, wantedPosMarked: []int{2}},

		{name: "First in sentenceq", data: &synthesizer.TTSData{Cfg: synthesizer.TTSConfig{JustAM: false},
			Words: []*synthesizer.ProcessedWord{
				{Tagged: synthesizer.TaggedWord{Word: "A"}},
				{Tagged: synthesizer.TaggedWord{Separator: ","}},
				{Tagged: synthesizer.TaggedWord{SentenceEnd: true}},
				{Tagged: synthesizer.TaggedWord{Word: "B"}},
				{Tagged: synthesizer.TaggedWord{Word: "olia"}},
			}}, wantedPosMarked: []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p ner
			gotErr := p.Process(t.Context(), tt.data)
			if gotErr != nil {
				t.Errorf("Process() failed: %v", gotErr)
				return
			}
			testValidateNER(t, tt.data.Words, tt.wantedPosMarked)
		})
	}
}

func testValidateNER(t testing.TB, words []*synthesizer.ProcessedWord, wantedPosMarked []int) {
	t.Helper()

	m := make(map[int]struct{})
	for _, v := range wantedPosMarked {
		m[v] = struct{}{}
	}

	for i, w := range words {
		_, wanted := m[i]
		if wanted {
			if w.NERType != synthesizer.NERSingleLetter {
				t.Errorf("word %s (%d) expected to be marked as NERSingleLetter, got %v", w.Tagged.Word, i, w.NERType)
			}
		} else {
			if w.NERType == synthesizer.NERSingleLetter {
				t.Errorf("word %s (%d) expected NOT to be marked as NERSingleLetter, got %v", w.Tagged.Word, i, w.NERType)
			}
		}
	}
}
