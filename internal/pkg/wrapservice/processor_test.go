package wrapservice

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/wrapservice/api"
)

var (
	httpAMMock  *mocks.MockHTTPInvokerJSON
	httpVocMock *mocks.MockHTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	mocks.AttachMockToTest(t)
	httpAMMock = mocks.NewMockHTTPInvokerJSON()
	httpVocMock = mocks.NewMockHTTPInvokerJSON()
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

	pegomock.When(httpAMMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*output) = output{Data: "specs"}
			return []pegomock.ReturnValue{nil}
		})
	pegomock.When(httpVocMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*output) = output{Data: "audio"}
			return []pegomock.ReturnValue{nil}
		})
	text, err := pr.Work(&api.Params{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10})
	assert.Nil(t, err)
	assert.Equal(t, "audio", text)

	cp1, _ := httpAMMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).GetCapturedArguments()
	assert.Equal(t, &amInput{Text: "olia", Speed: 0.9, Voice: "voice", Priority: 10}, cp1)

	cp2, _ := httpVocMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).GetCapturedArguments()
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

	pegomock.When(httpAMMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	pegomock.When(httpVocMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*output) = output{Data: "audio"}
			return []pegomock.ReturnValue{nil}
		})
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

	pegomock.When(httpAMMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*output) = output{Data: "audio"}
			return []pegomock.ReturnValue{nil}
		})
	pegomock.When(httpVocMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	_, err := pr.Work(&api.Params{Text: "olia", Speed: 1, Voice: "voice", Priority: 10})
	assert.NotNil(t, err)
	assert.InDelta(t, 1.0, testutil.ToFloat64(totalFailureMetrics.WithLabelValues("vocoder", "voice")), 0.000001)
}
