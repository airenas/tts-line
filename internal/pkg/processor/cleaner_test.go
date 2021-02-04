package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewCleaner(t *testing.T) {
	initTestJSON(t)
	pr, err := NewCleaner("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewCleaner_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAccentuator("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestCleanProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*normData) = normData{Text: "clean text"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	cp1, _ := httpJSONMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).GetCapturedArguments()
	assert.Equal(t, &normData{Text: " a a"}, cp1)
	assert.Equal(t, "clean text", d.Text)
}

func TestCleanProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestCleanProcess_NoText(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*normData) = normData{Text: ""}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "no_text", d.ValidationFailures[0].Check.ID)
}

func TestClean_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewCleaner("http://server")
	err := pr.Process(d)
	assert.Nil(t, err)
}
