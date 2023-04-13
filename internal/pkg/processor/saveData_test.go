package processor

import (
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
	dbMock *mockSaverDB
)

func initTestDB(t *testing.T) {
	dbMock = &mockSaverDB{}
}

func TestNewSaver(t *testing.T) {
	initTestDB(t)
	pr, err := NewSaver(dbMock, utils.RequestOriginal)
	assert.NotNil(t, pr)
	assert.Nil(t, err)
}

func TestNewSaver_Fail(t *testing.T) {
	initTestDB(t)
	_, err := NewSaver(nil, utils.RequestOriginal)
	assert.NotNil(t, err)
}

func TestSave_Ignore(t *testing.T) {
	initTestDB(t)
	pr, _ := NewSaver(dbMock, utils.RequestOriginal)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{AllowCollectData: false}
	err := pr.Process(d)
	assert.Nil(t, err)
	dbMock.AssertNumberOfCalls(t, "Save", 0)
}

func TestSave_Call(t *testing.T) {
	initTestDB(t)
	pr, _ := NewSaver(dbMock, utils.RequestOriginal)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.RequestID = "olia"
	d.OriginalText = "tata"
	d.Input = &api.TTSRequestConfig{AllowCollectData: true, SaveTags: []string{"olia"}}
	dbMock.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := pr.Process(d)
	assert.Nil(t, err)
	dbMock.AssertNumberOfCalls(t, "Save", 1)
	cRID := mocks.To[string](dbMock.Calls[0].Arguments[0])
	cText := mocks.To[string](dbMock.Calls[0].Arguments[1])
	cType := mocks.To[utils.RequestTypeEnum](dbMock.Calls[0].Arguments[2])
	cTags := mocks.To[[]string](dbMock.Calls[0].Arguments[3])
	assert.Equal(t, "olia", cRID)
	assert.Equal(t, "tata", cText)
	assert.Equal(t, utils.RequestOriginal, cType)
	assert.Equal(t, []string{"olia"}, cTags)
}

func TestSave_Normalized(t *testing.T) {
	initTestDB(t)
	pr, _ := NewSaver(dbMock, utils.RequestNormalized)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.RequestID = "olia"
	d.TextWithNumbers = "normalized"
	d.Input = &api.TTSRequestConfig{AllowCollectData: true}
	dbMock.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := pr.Process(d)
	assert.Nil(t, err)
	dbMock.AssertNumberOfCalls(t, "Save", 1)
	cRID := mocks.To[string](dbMock.Calls[0].Arguments[0])
	cText := mocks.To[string](dbMock.Calls[0].Arguments[1])
	cType := mocks.To[utils.RequestTypeEnum](dbMock.Calls[0].Arguments[2])
	assert.Equal(t, "olia", cRID)
	assert.Equal(t, "normalized", cText)
	assert.Equal(t, utils.RequestNormalized, cType)
}

func TestSave_Fail(t *testing.T) {
	initTestDB(t)
	pr, _ := NewSaver(dbMock, utils.RequestOriginal)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.RequestID = "olia"
	d.OriginalText = "tata"
	d.Input = &api.TTSRequestConfig{AllowCollectData: true}
	dbMock.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestGetText(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.RequestID = "olia"
	d.OriginalText = "tata"
	d.Text = "cleaned"
	d.TextWithNumbers = "t numbers"
	assert.Equal(t, "tata", getText(d, utils.RequestOriginal))
	assert.Equal(t, "cleaned", getText(d, utils.RequestCleaned))
	assert.Equal(t, "t numbers", getText(d, utils.RequestNormalized))
	assert.Equal(t, "tata", getText(d, utils.RequestUser))
	assert.Equal(t, "tata", getText(d, utils.RequestOriginalSSML))
}

type mockSaverDB struct{ mock.Mock }

func (m *mockSaverDB) Save(req, text string, reqType utils.RequestTypeEnum, tags []string) error {
	args := m.Called(req, text, reqType, tags)
	return args.Error(0)
}
