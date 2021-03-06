package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

var (
	httpJSONMock *mocks.MockHTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	mocks.AttachMockToTest(t)
	httpJSONMock = mocks.NewMockHTTPInvokerJSON()
}

func TestNewAbbreviator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcronyms("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewAbbreviator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcronyms("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeNewAbbreviator(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcronyms("http://server")
	assert.NotNil(t, pr)
	pr.(*acronyms).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Mi: "Y"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]acrWordOutput) = []acrWordOutput{{ID: "0",
				Words: []acrResultWord{{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "o l i a", d.Words[0].UserTranscription)
	assert.Equal(t, "o-lia", d.Words[0].UserSyllables)
	assert.Equal(t, "olia", d.Words[0].Tagged.Word)
}

func TestInvokeNewAbbreviator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcronyms("http://server")
	assert.NotNil(t, pr)
	pr.(*acronyms).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Mi: "Y"}})
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}
func TestInvokeAbbr_Skip(t *testing.T) {
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	pr, _ := NewAcronyms("http://server")
	err := pr.Process(d)
	assert.Nil(t, err)
}

func TestMapAbbrOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "0", Words: []acrResultWord{
		{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.Nil(t, err)
	assert.Equal(t, ",", d.Words[2].Tagged.Separator)
	assert.Equal(t, "o l i a", d.Words[0].UserTranscription)
	assert.Equal(t, "olia", d.Words[0].Tagged.Word)
}

func TestMapAbbrOutput_Several(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "1", Words: []acrResultWord{
		{Word: "v1", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"},
		{Word: "v2", WordTrans: "oolia", UserTrans: "v 2", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.Nil(t, err)
	assert.Equal(t, ",", d.Words[0].Tagged.Separator)
	assert.Equal(t, "v1", d.Words[1].Tagged.Word)
	assert.Equal(t, "v2", d.Words[2].Tagged.Word)
}

func TestMapAbbrOutput_Fail(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "XX", Words: []acrResultWord{
		{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.NotNil(t, err)
}

func TestIsAbbr(t *testing.T) {
	assert.True(t, isAbbr("Ya", "O"))
	assert.True(t, isAbbr("X", "O"))
	assert.True(t, isAbbr("N", "PP"))
	assert.False(t, isAbbr("Npmsnng", "Kaunas"))
	assert.False(t, isAbbr("Npmsnng", "Pp"))
}

func newTestTTSDataPart() *synthesizer.TTSDataPart {
	return &synthesizer.TTSDataPart{Cfg: &synthesizer.TTSConfig{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}}
}
