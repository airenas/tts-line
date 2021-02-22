package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewConverter(t *testing.T) {
	initTestJSON(t)
	pr, err := NewConverter("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewConverter_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewConverter("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeConvert(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Audio = "wav"
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*audioConvertOutput) = audioConvertOutput{Data: "mp3"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", d.AudioMP3)
	cData, _ := httpJSONMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).
		GetCapturedArguments()
	cInp, _ := cData.(audioConvertInput)
	assert.Equal(t, "wav", cInp.Data)
	assert.Equal(t, "mp3", cInp.Format)
	assert.Equal(t, []string{"olia"}, cInp.Metadata)
}

func TestInvokeConvert_Skip(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Audio = "wav"
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioNone}
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*audioConvertOutput) = audioConvertOutput{Data: "mp3"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	httpJSONMock.VerifyWasCalled(pegomock.Never()).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
}

func TestInvokeConvert_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{OutputFormat: api.AudioMP3}
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}
