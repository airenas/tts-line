package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
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
func TestSplit_NoSplitl(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	r, err := split(d.Words, 20)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r))
}

func TestSplit_OnSentenceEnd(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	r, err := split(d.Words, 25)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 3, len(r[0].Words))
}

func TestSplit_OnPunct(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	r, err := split(d.Words, 25)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 3, len(r[0].Words))
	assert.Equal(t, 2, len(r[1].Words))
}

func TestSplit_OnSpace(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	r, err := split(d.Words, 25)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 2, len(r[0].Words))
	assert.Equal(t, 2, len(r[1].Words))
}

func TestSplit_OnSpace2(t *testing.T) {
	d := synthesizer.TTSData{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "0123456789"}})
	r, err := split(d.Words, 25)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, 2, len(r[0].Words))
	assert.Equal(t, 1, len(r[1].Words))
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
