package processor

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]accentOutputElement) = []accentOutputElement{{Word: "word",
				Accent: []accentInfo{{Mi: "mi", Variants: []synthesizer.AccentVariant{{Accent: 101}}}}}}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
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
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]accentOutputElement) = []accentOutputElement{}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestInvokeAccentuator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	pr.(*accentuator).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestInvokeAccentuator_NoData(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAccentuator("http://server")
	assert.NotNil(t, pr)
	d := newTestTTSDataPart()
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
}

func TestInvokeAccentuator_Skip(t *testing.T) {
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	pr, _ := NewAccentuator("http://server")
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
}

func TestMapAccInput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v3"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	inp := mapAccentInput(d)
	assert.Equal(t, []string{"v2", "v3"}, inp)
}

func TestMapAccOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{{Mi: "mi", Variants: []synthesizer.AccentVariant{{Accent: 101,
			Syll: "v-1"}}}}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, 101, d.Words[3].AccentVariant.Accent)
	assert.Equal(t, "v-1", d.Words[3].AccentVariant.Syll)
}

func TestMapAccOutput_FindBest(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{UserTranscription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{{Accent: 101,
			Syll: "v-1"}}},
			{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{{Accent: 102,
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

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{{MiVdu: "mi1", Error: "err", Variants: []synthesizer.AccentVariant{{Accent: 0,
			Syll: "v-1"}}},
			{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{{Accent: 102,
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

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{
			{MiVdu: "mi1", Error: "err", Variants: []synthesizer.AccentVariant{{Accent: 0, Syll: "v-1"}}},
			{MiVdu: "mi2", Variants: []synthesizer.AccentVariant{
				{Accent: 0, Syll: "v-2"},
				{Accent: 103, Syll: "v-3"},
			}},
		}}}

	err := mapAccentOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "v-3", d.Words[2].AccentVariant.Syll)
	assert.Equal(t, 103, d.Words[2].AccentVariant.Accent)
}

func TestMapAccOutput_FailError(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{
			{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{{Accent: 0, Syll: "v-1"}}},
		}},
		{Word: "v2", Error: "error olia"}}

	err := mapAccentOutput(d, output)
	if assert.NotNil(t, err) {
		assert.Equal(t, "accent error for 'v2'('v2'): error olia", err.Error())
	}
}

func TestMapAccOutput_FailErrorTooLong(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{
		Word: "loooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong", Mi: "mi2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{
			{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{{Accent: 0, Syll: "v-1"}}},
		}},
		{Word: "v2", Error: "error olia"}}

	err := mapAccentOutput(d, output)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "wrong accent, too long word: ")
	}
}

func TestMapAccOutput_FailNoWordAccented(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{
		Word: "bad", Mi: "mi2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{
			{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{{Accent: 0, Syll: "v-1"}}},
		}},
		{Word: "v2", Error: "No word"}}

	err := mapAccentOutput(d, output)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "wrong symbols: 'bad' ()")
	}
}

func TestMapAccOutput_FailWrongWord(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v2", Mi: "mi2"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{
		Word: "badd", Mi: "mi2"}})

	output := []accentOutputElement{{Word: "v2",
		Accent: []accentInfo{
			{MiVdu: "mi1", Variants: []synthesizer.AccentVariant{{Accent: 0, Syll: "v-1"}}},
		}},
		{Word: "bad", Error: ""}}

	err := mapAccentOutput(d, output)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "wrong symbols: 'badd' (bad)")
	}
}

func TestFindBest_UseLemma(t *testing.T) {
	acc := []accentInfo{{MiVdu: "mi2", MF: "lema1", Variants: []synthesizer.AccentVariant{{Accent: 101}}},
		{MiVdu: "mi2", MF: "lema", Variants: []synthesizer.AccentVariant{{Accent: 103}}}}
	res := findBestAccentVariant(acc, "mi2", "lema")

	assert.Equal(t, 103, res.Accent)
}
