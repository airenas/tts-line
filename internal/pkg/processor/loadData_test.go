package processor

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

var (
	loadMock *mockLoadDB
)

func initLoadTestDB(t *testing.T) {
	loadMock = &mockLoadDB{}
}

func TestNewLoader(t *testing.T) {
	initLoadTestDB(t)
	pr, err := NewLoader(loadMock)
	assert.NotNil(t, pr)
	assert.Nil(t, err)
}

func TestNewLoader_Fail(t *testing.T) {
	initLoadTestDB(t)
	_, err := NewLoader(nil)
	assert.NotNil(t, err)
}

func TestLoad(t *testing.T) {
	initLoadTestDB(t)
	pr, _ := NewLoader(loadMock)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{RequestID: "i1"}
	loadMock.On("LoadText", mock.Anything, mock.Anything).Return("olia", nil)

	err := pr.Process(context.TODO(), d)

	assert.Nil(t, err)
	assert.Equal(t, "olia", d.PreviousText)

	loadMock.AssertNumberOfCalls(t, "LoadText", 1)
	cr := mocks.To[string](loadMock.Calls[0].Arguments[0])
	ct := mocks.To[utils.RequestTypeEnum](loadMock.Calls[0].Arguments[1])

	assert.Equal(t, "i1", cr)
	assert.Equal(t, utils.RequestNormalized.String(), ct.String())
}

func TestLoad_Fail(t *testing.T) {
	initLoadTestDB(t)
	pr, _ := NewLoader(loadMock)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{RequestID: "i1"}
	loadMock.On("LoadText", mock.Anything, mock.Anything).Return("", errors.New("olia"))

	err := pr.Process(context.TODO(), d)

	assert.NotNil(t, err)
}

type mockLoadDB struct{ mock.Mock }

func (m *mockLoadDB) LoadText(req string, reqType utils.RequestTypeEnum) (string, error) {
	args := m.Called(req, reqType)
	return mocks.To[string](args.Get(0)), args.Error(1)
}
