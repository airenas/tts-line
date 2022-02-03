package utils_test

import (
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

var (
	testHTTPWrap *mocks.MockHTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	mocks.AttachMockToTest(t)
	testHTTPWrap = mocks.NewMockHTTPInvokerJSON()
}

func TestNewHTTPBackoff(t *testing.T) {
	initTestJSON(t)
	pr, err := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff { return backoff.NewExponentialBackOff() }, utils.RetryAll)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestInvokeJSON(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff { return backoff.NewExponentialBackOff() }, utils.RetryAll)
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(nil)

	err := pr.InvokeJSON("olia", "")
	assert.Nil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(1))
}

func TestInvokeRetry(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
	}, utils.RetryAll)
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))

	err := pr.InvokeJSON("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(4))
}

func TestCallbacks(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
	}, utils.RetryAll)
	ic := 0
	rc := 0
	pr.InvokeIndicatorFunc = func(d interface{}) {
		ic++
		assert.Equal(t, "olia", d)
	}
	pr.RetryIndicatorFunc = func(d interface{}) {
		rc++
		assert.Equal(t, "olia", d)
	}
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))

	err := pr.InvokeJSON("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(4))
	assert.Equal(t, 4, ic)
	assert.Equal(t, 3, rc)
}
