package cache

import (
	"strings"
	"testing"
	"time"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/test/mocks/matchers"
)

var (
	synthesizerMock *mocks.MockSynthesizer
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	synthesizerMock = mocks.NewMockSynthesizer()
}

func TestNewCacher(t *testing.T) {
	initTest(t)
	c, err := NewCacher(synthesizerMock, newTestConfig(""))
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Nil(t, c.cache)
}

func TestNewCacher_Fails(t *testing.T) {
	initTest(t)
	_, err := NewCacher(nil, newTestConfig(""))
	assert.NotNil(t, err)
}

func TestNewCacherInit(t *testing.T) {
	initTest(t)
	c, err := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.cache)
}

func TestCleanDuartion(t *testing.T) {
	assert.Equal(t, 5*time.Minute, getCleanDuration(0))
	assert.Equal(t, 2*time.Second, getCleanDuration(2*time.Second))
}

func TestWork(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	res, err := c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	assert.Equal(t, "wav", res.AudioAsString)
	synthesizerMock.VerifyWasCalledOnce().Work(matchers.AnyPtrToApiTTSRequestConfig())
	res, err = c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	assert.Equal(t, "wav", res.AudioAsString)
	synthesizerMock.VerifyWasCalledOnce().Work(matchers.AnyPtrToApiTTSRequestConfig())
	res, err = c.Work(newtestInput("olia2"))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
}

func TestWork_Failure(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(nil, errors.New("haha"))

	_, err := c.Work(newtestInput("olia"))
	assert.NotNil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(matchers.AnyPtrToApiTTSRequestConfig())
	_, err = c.Work(newtestInput("olia"))
	assert.NotNil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
}

func TestWork_NoCache(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 0s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	_, err := c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(matchers.AnyPtrToApiTTSRequestConfig())
	_, err = c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
}

func TestWork_NoCacheValidation(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{
			ValidationFailures: []api.ValidateFailure{api.ValidateFailure{}}}, nil)

	_, err := c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(matchers.AnyPtrToApiTTSRequestConfig())
	_, err = c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
}

func TestWork_Key(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(matchers.AnyPtrToApiTTSRequestConfig())).
		ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	c.Work(newtestInput("olia"))
	c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3})
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
	c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3})
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(matchers.AnyPtrToApiTTSRequestConfig())
	c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioM4A})
	synthesizerMock.VerifyWasCalled(pegomock.Times(3)).Work(matchers.AnyPtrToApiTTSRequestConfig())
}

func Test_Key(t *testing.T) {
	initTest(t)
	assert.Equal(t, "olia_mp3_", key(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3}))
	assert.Equal(t, "olia1_m4a_accented", key(&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioM4A,
		OutputTextFormat: api.TextAccented}))
}

func newTestConfig(yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	res.ReadConfig(strings.NewReader(yaml))
	return res
}

func newtestInput(txt string) *api.TTSRequestConfig {
	return &api.TTSRequestConfig{Text: txt}
}
