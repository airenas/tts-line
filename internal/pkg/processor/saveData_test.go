package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/test/mocks/matchers"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

var (
	dbMock *mocks.MockSaverDB
)

func initTestDB(t *testing.T) {
	mocks.AttachMockToTest(t)
	dbMock = mocks.NewMockSaverDB()
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
	dbMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), pegomock.AnyString(),
		matchers.AnyUtilsRequestTypeEnum(), pegomock.AnyStringSlice())
}

func TestSave_Call(t *testing.T) {
	initTestDB(t)
	pr, _ := NewSaver(dbMock, utils.RequestOriginal)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.RequestID = "olia"
	d.OriginalText = "tata"
	d.Input = &api.TTSRequestConfig{AllowCollectData: true, SaveTags: []string{"olia"}}
	err := pr.Process(d)
	assert.Nil(t, err)
	cRID, cText, cType, cTags := dbMock.VerifyWasCalled(pegomock.Once()).
		Save(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyUtilsRequestTypeEnum(),
			pegomock.AnyStringSlice()).
		GetCapturedArguments()
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
	err := pr.Process(d)
	assert.Nil(t, err)
	cRID, cText, cType, _ := dbMock.VerifyWasCalled(pegomock.Once()).
		Save(pegomock.AnyString(), pegomock.AnyString(), matchers.AnyUtilsRequestTypeEnum(),
			pegomock.AnyStringSlice()).
		GetCapturedArguments()
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
	pegomock.When(dbMock.Save(pegomock.AnyString(), pegomock.AnyString(),
		matchers.AnyUtilsRequestTypeEnum(), pegomock.AnyStringSlice())).
		ThenReturn(errors.New("haha"))
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
}
