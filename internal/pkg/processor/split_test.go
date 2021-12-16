package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
)

func TestNewSplit(t *testing.T) {
	pr := NewSplitter(30)
	assert.NotNil(t, pr)
	assert.Equal(t, 30, pr.(*splitter).maxChars)
}

func TestNewSplit_Default(t *testing.T) {
	pr := NewSplitter(0)
	assert.NotNil(t, pr)
	assert.Equal(t, 400, pr.(*splitter).maxChars)
}

func TestSplitter(t *testing.T) {
	pr := NewSplitter(0)
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(d.Parts))
}

func TestSplitter_NoData(t *testing.T) {
	pr := NewSplitter(0)
	d := synthesizer.TTSData{}

	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(d.Parts))
}

func TestSplitter_Fail(t *testing.T) {
	pr := NewSplitter(9)
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestSplit_Several(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	r, err := split(d.Words, 10)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, d.Words[0], r[0].Words[0])
	assert.Equal(t, d.Words[1], r[0].Words[1])
	assert.Equal(t, d.Words[2], r[1].Words[0])
	assert.Equal(t, d.Words[3], r[1].Words[1])
	assert.Equal(t, d.Words[4], r[1].Words[2])
}

func Test_split(t *testing.T) {
	type args struct {
		data []synthesizer.TaggedWord
		max  int
	}
	tests := []struct {
		name     string
		args     args
		wantLens []int
		wantErr  bool
	}{
		{name: "NoSplit", args: args{max: 20,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {SentenceEnd: true},
				{Word: "0123456789"}, {Separator: "?"}, {SentenceEnd: true}}},
			wantLens: []int{5}, wantErr: false},
		{name: "Split", args: args{max: 19,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {SentenceEnd: true},
				{Word: "0123456789"}, {Separator: "?"}, {SentenceEnd: true}}},
			wantLens: []int{2, 3}, wantErr: false},
		{name: "Fail", args: args{max: 9,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {SentenceEnd: true}}},
			wantLens: []int{}, wantErr: true},
		{name: "OnSentenceEnd", args: args{max: 25,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {Word: "0123456789"}, {SentenceEnd: true},
				{Word: "0123456789"}, {Separator: "?"}}},
			wantLens: []int{3, 2}, wantErr: false},
		{name: "OnPunctSpace", args: args{max: 25,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {Separator: ","}, {Space: true},
				{Word: "0123456789"}, {Separator: ","},
				{Word: "0123456789"}, {Separator: "?"}}},
			wantLens: []int{3, 4}, wantErr: false},
		{name: "OnPunct", args: args{max: 25,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {Word: "0123456789"}, {Separator: ","},
				{Word: "0123456789"}, {Separator: "?"}}},
			wantLens: []int{3, 2}, wantErr: false},
		{name: "Ignore Empty", args: args{max: 12,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {SentenceEnd: true},
				{Space: true}, {Separator: ","}, {Space: true}, {Word: "0123456789"}, {Space: true},
				{Word: "0123456789"}}},
			wantLens: []int{2, 5, 1}, wantErr: false},
		{name: "OnSpace", args: args{max: 25,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {Word: "0123456789"}, {Space: true},
				{Word: "0123456789"}, {Separator: "?"}}},
			wantLens: []int{3, 2}, wantErr: false},
		{name: "OnSpace2", args: args{max: 25,
			data: []synthesizer.TaggedWord{
				{Word: "0123456789"}, {Space: true}, {Word: "0123456789"}, {Space: true},
				{Word: "0123456789"}, {Space: true}}},
			wantLens: []int{4, 2}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := []*synthesizer.ProcessedWord{}
			for _, t := range tt.args.data {
				d = append(d, &synthesizer.ProcessedWord{Tagged: t})
			}
			got, err := split(d, tt.args.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if assert.Equal(t, len(tt.wantLens), len(got), "fails len") {
				for i, wl := range tt.wantLens {
					assert.Equal(t, wl, len(got[i].Words), "word len failing for got[%d].Words", i)
				}
			}
		})
	}
}

func TestSplit_First(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	r, err := split(d.Words, 15)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(r))
	assert.True(t, r[0].First)
	assert.False(t, r[1].First)
	assert.False(t, r[2].First)
}
