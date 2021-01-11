package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewAccentuator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAccentuator("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewAccentuator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAccentuator("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeAccentuator(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	pr.(*accentuator).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]accentOutputElement) = []accentOutputElement{accentOutputElement{Word: "word",
				Accent: []accent{accent{Mi: "mi", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 101}}}}}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, 101, d.Words[0].AccentVariant.Accent)
}

func TestInvokeAccentuator_FailOutput(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	pr.(*accentuator).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]accentOutputElement) = []accentOutputElement{}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestInvokeAccentuator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	pr.(*accentuator).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestInvokeAccentuator_NoData(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	d := newTestTTSDataPart()
	err := pr.Process(d)
	assert.Nil(t, err)
}

func TestMapAccInput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v3"}})
	inp := mapAccentInput(d)
	assert.Equal(t, []string{"v2", "v3"}, inp)
}

func TestMapAccOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2"}})

	output := []accentOutputElement{accentOutputElement{Word: "v2",
		Accent: []accent{accent{Mi: "mi", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 101,
			Syll: "v-1"}}}}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, 101, d.Words[2].AccentVariant.Accent)
	assert.Equal(t, "v-1", d.Words[2].AccentVariant.Syll)
}

func TestMapAccOutput_FindBest(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})

	output := []accentOutputElement{accentOutputElement{Word: "v2",
		Accent: []accent{accent{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 101,
			Syll: "v-1"}}},
			accent{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 102,
				Syll: "v-1"}}},
		}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, 102, d.Words[2].AccentVariant.Accent)
}

func TestMapAccOutput_Error(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})

	output := []accentOutputElement{accentOutputElement{Word: "v2",
		Accent: []accent{accent{MiVdu: "mi1", Error: "err", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 0,
			Syll: "v-1"}}},
			accent{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 102,
				Syll: "v-2"}}},
		}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "v-2", d.Words[2].AccentVariant.Syll)
}

func TestMapAccOutput_WithAccent(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})

	output := []accentOutputElement{accentOutputElement{Word: "v2",
		Accent: []accent{accent{MiVdu: "mi1", Error: "err", Variants: []synthesizer.AccentVariant{synthesizer.AccentVariant{Accent: 0,
			Syll: "v-1"}}},
			accent{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{
				synthesizer.AccentVariant{Accent: 0, Syll: "v-2"},
				synthesizer.AccentVariant{Accent: 103, Syll: "v-3"},
			}},
		}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "v-3", d.Words[2].AccentVariant.Syll)
	assert.Equal(t, 103, d.Words[2].AccentVariant.Accent)
}
