package wrapservice

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/airenas/tts-line/internal/pkg/syntmodel"
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
			*params[1].(*syntmodel.AMOutput) = syntmodel.AMOutput{Data: []byte("specs"), Durations: []int{10, 12}, SilDuration: 15}
		}).Return(nil)

	httpVocMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*syntmodel.VocOutput) = syntmodel.VocOutput{Data: []byte("audio")}
		}).Return(nil)
	res, err := pr.Work(context.TODO(), &api.Params{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10})
	assert.Nil(t, err)
	assert.Equal(t, &syntmodel.Result{Data: []byte("audio"), Durations: []int{10, 12}, SilDuration: 15}, res)

	httpAMMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp1 := httpAMMock.Calls[0].Arguments[0]
	assert.Equal(t, &syntmodel.AMInput{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10}, cp1)

	httpVocMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp2 := httpVocMock.Calls[0].Arguments[0]
	assert.Equal(t, &syntmodel.VocInput{Data: []byte("specs"), Voice: "voice", Priority: 10}, cp2)
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
			*params[1].(*syntmodel.VocOutput) = syntmodel.VocOutput{Data: []byte("audio")}
		}).Return(nil)
	_, err := pr.Work(context.TODO(), &api.Params{Text: "olia", Speed: 1, Voice: "voice", Priority: 10})
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
			*params[1].(*syntmodel.AMOutput) = syntmodel.AMOutput{Data: []byte("audio")}
		}).Return(nil)
	httpVocMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	_, err := pr.Work(context.TODO(), &api.Params{Text: "olia", Speed: 1, Voice: "voice", Priority: 10})
	assert.NotNil(t, err)
	assert.InDelta(t, 1.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("vocoder", "voice")), 0.000001)
}
