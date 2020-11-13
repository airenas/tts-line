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
	pegomock.When(synthesizerMock.Work(pegomock.AnyString())).ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	res, err := c.Work("olia")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	synthesizerMock.VerifyWasCalledOnce().Work(pegomock.AnyString())
	res, err = c.Work("olia")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	synthesizerMock.VerifyWasCalledOnce().Work(pegomock.AnyString())
	res, err = c.Work("olia2")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(pegomock.AnyString())
}

func TestWork_Failure(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(pegomock.AnyString())).ThenReturn(nil, errors.New("haha"))

	_, err := c.Work("olia")
	assert.NotNil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(pegomock.AnyString())
	_, err = c.Work("olia")
	assert.NotNil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(pegomock.AnyString())
}

func TestWork_NoCache(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 0s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(pegomock.AnyString())).ThenReturn(&api.Result{AudioAsString: "wav"}, nil)

	_, err := c.Work("olia")
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(pegomock.AnyString())
	_, err = c.Work("olia")
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(pegomock.AnyString())
}

func TestWork_NoCacheValidation(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	pegomock.When(synthesizerMock.Work(pegomock.AnyString())).ThenReturn(&api.Result{
		ValidationFailures: []api.ValidateFailure{api.ValidateFailure{}}}, nil)

	_, err := c.Work("olia")
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalledOnce().Work(pegomock.AnyString())
	_, err = c.Work("olia")
	assert.Nil(t, err)
	synthesizerMock.VerifyWasCalled(pegomock.Twice()).Work(pegomock.AnyString())
}

func newTestConfig(yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	res.ReadConfig(strings.NewReader(yaml))
	return res
}
