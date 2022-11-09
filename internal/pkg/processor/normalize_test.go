package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewNormalizer(t *testing.T) {
	initTestJSON(t)
	pr, err := NewNormalizer("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewNormalizer_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewNormalizer("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestNormalizeProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.CleanedText = " a a"
	pr, _ := NewNormalizer("http://server")
	assert.NotNil(t, pr)
	pr.(*normalizer).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*normResponseData) = normResponseData{Res: "out text"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	cp1, _ := httpJSONMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).GetCapturedArguments()
	assert.Equal(t, &normRequestData{Orig: " a a"}, cp1)
	assert.Equal(t, "out text", d.NormalizedText)
}

func TestNormalizeProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.CleanedText = " a a"
	pr, _ := NewNormalizer("http://server")
	assert.NotNil(t, pr)
	pr.(*normalizer).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestNormalize_Skip(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewNormalizer("http://server")
	pr.(*normalizer).httpWrap = httpJSONMock
	err := pr.Process(d)
	assert.Nil(t, err)
	httpJSONMock.VerifyWasCalled(pegomock.Never()).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
}
