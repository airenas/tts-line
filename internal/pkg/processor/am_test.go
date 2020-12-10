package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewAcousticModel(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcousticModel("http://server", "")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewAcousticModel_Space(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel("http://server", "")
	assert.NotNil(t, pr)
	assert.Equal(t, "sil", pr.(*amodel).spaceSymbol)
	pr, _ = NewAcousticModel("http://server", "<space>")
	assert.NotNil(t, pr)
	assert.Equal(t, "<space>", pr.(*amodel).spaceSymbol)
}

func TestNewAcousticModel_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcousticModel("", "")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeAcousticModel(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel("http://server", "")
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := synthesizer.TTSDataPart{}
	d.Spectogram = "spectogram"
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*amOutput) = amOutput{Data: "spec"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "spec", d.Spectogram)
}

func TestInvokeAcousticModel_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel("http://server", "")
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := synthesizer.TTSDataPart{}
	d.Spectogram = "spectogram"
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestMapAMInput(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := synthesizer.TTSDataPart{First: true}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "<space> v a o l i a , v a <space>", inp.Text)
}

func TestMapAMInput_NoSilAtStart(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := synthesizer.TTSDataPart{First: false}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "v a o l i a , v a sil", inp.Text)
}

func TestMapAMInput_SpaceDot(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := synthesizer.TTSDataPart{First: true}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "<space> v a o l i a . <space> v a <space>", inp.Text)
}

func TestMapAMInput_SpaceQuestion(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := synthesizer.TTSDataPart{First: true}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "<space> v a o l i a ? <space> v a <space>", inp.Text)
}

func TestMapAMInput_Exclamation(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := synthesizer.TTSDataPart{First: true}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "<space> v a o l i a ! <space> v a <space>", inp.Text)
}

func TestMapAMInput_SpaceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := synthesizer.TTSDataPart{First: true}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "<space> v a <space>", inp.Text)
}

func TestMapAMInput_AddDotOnSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := synthesizer.TTSDataPart{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "v a . sil", inp.Text)
}

func TestMapAMInput_NoDotOnSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := synthesizer.TTSDataPart{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "v a . sil", inp.Text)
}

func TestMapAMInput_NoDotOnSentenceEnd2(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := synthesizer.TTSDataPart{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "v a ? sil", inp.Text)
}

func TestMapAMInput_SeveralSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := synthesizer.TTSDataPart{}
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp := pr.mapAMInput(&d)
	assert.Equal(t, "v a . sil v a ? sil", inp.Text)
}

func TestSep(t *testing.T) {
	for _, s := range ",:.?!-" {
		assert.Equal(t, string(s), getSep(string(s)))
	}
	assert.Equal(t, "...", getSep("..."))
	assert.Equal(t, ",", getSep(";"))
	assert.Equal(t, "", getSep("\""))
}

func newTestAM(t *testing.T, urlStr string, spaceSym string) *amodel {
	pr, err := NewAcousticModel(urlStr, spaceSym)
	assert.Nil(t, err)
	return pr.(*amodel)
}
