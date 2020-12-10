package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewMP3(t *testing.T) {
	initTestJSON(t)
	pr, err := NewMP3("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewMP3_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewMP3("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeMP3(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewMP3("http://server")
	assert.NotNil(t, pr)
	pr.(*mp3Converter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Audio = "wav"
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*mp3Output) = mp3Output{Data: "mp3"}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", d.AudioMP3)
}

func TestInvokeMP3_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewMP3("http://server")
	assert.NotNil(t, pr)
	pr.(*mp3Converter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}
