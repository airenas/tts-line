package synthesizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

var (
	processorMock *procMock
	worker        *MainWorker
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	processorMock = &procMock{f: func(d *TTSData) error { return nil }}
	worker = &MainWorker{}
	worker.Add(processorMock)
}

func TestWork(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		assert.Equal(t, "olia", d.OriginalText)
		d.AudioMP3 = "mp3"
		return nil
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "mp3", res.AudioAsString)
}

func TestWork_Fails(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		return errors.New("olia")
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWork_ValidationFailure(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.ValidationFailures = append(d.ValidationFailures, api.ValidateFailure{FailingPosition: 10})
		return nil
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "", res.AudioAsString)
	assert.Equal(t, 10, res.ValidationFailures[0].FailingPosition)
}

func TestWork_Several(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.AudioMP3 = "wav"
		return nil
	}
	processorMock1 := &procMock{f: func(d *TTSData) error {
		d.AudioMP3 = d.AudioMP3 + "mp3"
		return nil
	}}
	worker.Add(processorMock1)
	res, _ := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Equal(t, "wavmp3", res.AudioAsString)
}

func TestWork_StopProcess(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.ValidationFailures = append(d.ValidationFailures, api.ValidateFailure{FailingPosition: 10})
		return nil
	}
	processorMock1 := &procMock{f: func(d *TTSData) error {
		assert.Fail(t, "Unexpected call")
		return nil
	}}
	worker.Add(processorMock1)
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

type procMock struct {
	f func(res *TTSData) error
}

func (pr *procMock) Process(d *TTSData) error {
	return pr.f(d)
}
