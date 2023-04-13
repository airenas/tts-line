package wrapservice

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"
)

var (
	httpAMMock  *mocks.HTTPInvokerJSON
	httpVocMock *mocks.HTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	httpAMMock = &mocks.HTTPInvokerJSON{}
	httpVocMock = &mocks.HTTPInvokerJSON{}
}

func TestNewProcessor(t *testing.T) {
	initTestJSON(t)
	pr, err := NewProcessor("http://server", "http://service1")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewProcessor_Fail(t *testing.T) {
	initTestJSON(t)
	pr, err := NewProcessor("http://server", "")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
	pr, err = NewProcessor("", "http://server")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeProcessor(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewProcessor("http://server", "http://service1")
	assert.NotNil(t, pr)
	pr.amWrap = httpAMMock
	pr.vocWrap = httpVocMock

	httpAMMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*output) = output{Data: "specs"}
		}).Return(nil)

	httpVocMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*output) = output{Data: "audio"}
		}).Return(nil)
	text, err := pr.Work(&api.Params{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10})
	assert.Nil(t, err)
	assert.Equal(t, "audio", text)

	httpAMMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp1 := httpAMMock.Calls[0].Arguments[0]
	assert.Equal(t, &amInput{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10}, cp1)

	httpVocMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp2 := httpVocMock.Calls[0].Arguments[0]
	assert.Equal(t, &vocInput{Data: "specs", Voice: "voice", Priority: 10}, cp2)
	assert.InDelta(t, 0.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("am", "voice")), 0.000001)
	assert.InDelta(t, 0.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("vocoder", "voice")), 0.000001)
}

func TestInvokeProcessor_FailAM(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewProcessor("http://server", "http://service1")
	assert.NotNil(t, pr)
	pr.amWrap = httpAMMock
	pr.vocWrap = httpVocMock

	httpAMMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	httpVocMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*output) = output{Data: "audio"}
		}).Return(nil)
	_, err := pr.Work(&api.Params{Text: "olia", Speed: 1, Voice: "voice", Priority: 10})
	assert.NotNil(t, err)
	assert.InDelta(t, 1.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("am", "voice")), 0.000001)
}

func TestInvokeProcessor_FailVoc(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewProcessor("http://server", "http://service1")
	assert.NotNil(t, pr)
	pr.amWrap = httpAMMock
	pr.vocWrap = httpVocMock

	httpAMMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*output) = output{Data: "audio"}
		}).Return(nil)
	httpVocMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	_, err := pr.Work(&api.Params{Text: "olia", Speed: 1, Voice: "voice", Priority: 10})
	assert.NotNil(t, err)
	assert.InDelta(t, 1.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("vocoder", "voice")), 0.000001)
}
