package processor

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*audioConvertOutput) = audioConvertOutput{Data: "mp3"}
		}).Return(nil)
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", d.AudioMP3)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cData := httpJSONMock.Calls[0].Arguments[0]
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
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*audioConvertOutput) = audioConvertOutput{Data: "mp3"}
		}).Return(nil)
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)

	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 0)
}

func TestInvokeConvert_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).httpWrap = httpJSONMock
	d := synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{OutputFormat: api.AudioMP3}
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}
