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
	loadMock *mocks.MockLoadDB
)

func initLoadTestDB(t *testing.T) {
	mocks.AttachMockToTest(t)
	loadMock = mocks.NewMockLoadDB()
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
	pegomock.When(loadMock.LoadText(pegomock.AnyString(), matchers.AnyUtilsRequestTypeEnum())).
		ThenReturn("olia", nil)

	err := pr.Process(d)

	assert.Nil(t, err)
	assert.Equal(t, "olia", d.PreviousText)
	cr, ct := loadMock.VerifyWasCalled(pegomock.Once()).
		LoadText(pegomock.AnyString(), matchers.AnyUtilsRequestTypeEnum()).GetCapturedArguments()
	assert.Equal(t, "i1", cr)
	assert.Equal(t, utils.RequestCleaned, ct)
}

func TestLoad_Fail(t *testing.T) {
	initLoadTestDB(t)
	pr, _ := NewLoader(loadMock)
	assert.NotNil(t, pr)
	d := &synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{RequestID: "i1"}
	pegomock.When(loadMock.LoadText(pegomock.AnyString(), matchers.AnyUtilsRequestTypeEnum())).
		ThenReturn("", errors.New("olia"))

	err := pr.Process(d)

	assert.NotNil(t, err)
}
