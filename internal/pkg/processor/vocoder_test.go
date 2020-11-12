package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewVocoder(t *testing.T) {
	initTestJSON(t)
	pr, err := NewVocoder("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewVocoder_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewVocoder("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeVocoder(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewVocoder("http://server")
	assert.NotNil(t, pr)
	pr.(*vocoder).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Spectogram = "spectogram"
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(**vocOutput) = &vocOutput{Data: "wav"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "wav", d.Audio)
}

func TestInvokeVocoder_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewVocoder("http://server")
	assert.NotNil(t, pr)
	pr.(*vocoder).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Spectogram = "spectogram"
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}
