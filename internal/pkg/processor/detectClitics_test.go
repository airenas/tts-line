package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/clitics/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestNewClitics(t *testing.T) {
	initTestJSON(t)
	pr, err := NewClitics("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewClitics_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewClitics("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeClitics(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewClitics("http://server")
	assert.NotNil(t, pr)
	pr.(*cliticDetector).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]api.CliticsOutput) = []api.CliticsOutput{{ID: 0,
				Type: "CLITIC", AccentType: api.TypeNone}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, synthesizer.CliticsNone, d.Words[0].Clitic.Type)
}

func TestInvokeClitics_Skip(t *testing.T) {
	initTestJSON(t)
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	pr, _ := NewClitics("http://server")
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(
		errors.New("olia err"))
	err := pr.Process(d)
	assert.Nil(t, err)
}

func TestInvokeClitics_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewClitics("http://server")
	assert.NotNil(t, pr)
	pr.(*cliticDetector).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(
		errors.New("olia err"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestMapCliticsInput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "olia", Lemma: "lemma", Mi: "mi"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "Olia1", Lemma: "lemma", Mi: "mi"}})
	inp, err := mapCliticsInput(d)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(inp)) {
		assert.Equal(t, "olia", inp[0].String)
		assert.Equal(t, "mi", inp[0].Mi)
		assert.Equal(t, "lemma", inp[0].Lemma)
		assert.Equal(t, 0, inp[0].ID)
		assert.Equal(t, "olia1", inp[1].String)
		assert.Equal(t, 1, inp[1].ID)
	}
}

func TestMapCliticsOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "olia", Lemma: "lemma", Mi: "mi"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "olia1", Lemma: "lemma", Mi: "mi"}})

	output := []api.CliticsOutput{{ID: 0, Type: "CLITIC", Accent: 103, AccentType: api.TypeStatic},
		{ID: 1, Type: "PHRASE", AccentType: "NONE"}}

	err := mapCliticsOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, synthesizer.CliticsCustom, d.Words[0].Clitic.Type)
	assert.Equal(t, 103, d.Words[0].Clitic.Accent)
	assert.Equal(t, synthesizer.CliticsNone, d.Words[1].Clitic.Type)
}

func TestToType(t *testing.T) {
	assert.Equal(t, "OTHER", toType(&synthesizer.TaggedWord{SentenceEnd: true}))
	assert.Equal(t, "SPACE", toType(&synthesizer.TaggedWord{Space: true}))
	assert.Equal(t, "OTHER", toType(&synthesizer.TaggedWord{Separator: ","}))
	assert.Equal(t, "WORD", toType(&synthesizer.TaggedWord{Word: "v1"}))
}
