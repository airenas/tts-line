package processor

import (
	"context"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/syntmodel"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	pr, _ := NewVocoder("http://{{voice}}.server")
	assert.NotNil(t, pr)
	pr.(*vocoder).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Spectogram = []byte("spectogram")
	d.Cfg.Input.Voice = "aaa"
	d.Cfg.Input.Priority = 10
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[2].(*syntmodel.VocOutput) = syntmodel.VocOutput{Data: []byte("wav")}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, []byte("wav"), d.Audio)

	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSONU", 1)
	url := httpJSONMock.Calls[0].Arguments[0]
	inp := httpJSONMock.Calls[0].Arguments[1]
	ai := inp.(syntmodel.VocInput)
	assert.Equal(t, "aaa", ai.Voice)
	assert.Equal(t, "http://aaa.server", url)
	assert.Equal(t, 10, ai.Priority)
}

func TestInvokeVocoder_Skip(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewVocoder("http://server")
	assert.NotNil(t, pr)
	pr.(*vocoder).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Cfg.Input.OutputFormat = api.AudioNone
	d.Spectogram = []byte("spectogram")
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSONU", 0)
}

func TestInvokeVocoder_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewVocoder("http://server")
	assert.NotNil(t, pr)
	pr.(*vocoder).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Spectogram = []byte("spectogram")
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}
