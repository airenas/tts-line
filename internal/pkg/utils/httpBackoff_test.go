package utils_test

import (
	"context"
	"io"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	testHTTPWrap.VerifyWasCalled(pegomock.Times(1)).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
}

func TestInvokeText(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap,
		func() backoff.BackOff { return backoff.NewExponentialBackOff() },
		utils.RetryAll)
	pegomock.When(testHTTPWrap.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).ThenReturn(nil)

	err := pr.InvokeText("olia", "")
	assert.Nil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(1)).InvokeText(pegomock.AnyString(), pegomock.AnyInterface())
}

func TestInvokeText_Retry(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap,
		func() backoff.BackOff {
			return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
		}, utils.RetryAll)
	pegomock.When(testHTTPWrap.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))

	err := pr.InvokeText("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(4)).InvokeText(pegomock.AnyString(), pegomock.AnyInterface())
}

func TestInvokeRetry(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
	}, utils.RetryAll)
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))

	err := pr.InvokeJSON("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(4)).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
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
	testHTTPWrap.VerifyWasCalled(pegomock.Times(4)).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
	assert.Equal(t, 4, ic)
	assert.Equal(t, 3, rc)
}

func TestRetry_StopsNonEOF(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 4)
	}, utils.RetryEOF)
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))
	err := pr.InvokeJSON("olia", "")
	require.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(1)).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
}

func TestRetry_ContinueEOF(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 4)
	}, utils.RetryEOF)
	pegomock.When(testHTTPWrap.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(io.EOF)
	err := pr.InvokeJSON("olia", "")
	require.NotNil(t, err)
	testHTTPWrap.VerifyWasCalled(pegomock.Times(5)).InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())
}

func TestRetryEOF(t *testing.T) {
	type args struct {
		err error
	}
	sErr := errors.Errorf("olia")
	tests := []struct {
		name        string
		args        args
		wantPermErr bool
	}{
		{name: "simple", args: args{err: sErr}, wantPermErr: true},
		{name: "EOF", args: args{err: io.EOF}, wantPermErr: false},
		{name: "Timeout", args: args{err: context.DeadlineExceeded}, wantPermErr: false},
		{name: "Wrapped EOF", args: args{err: errors.Wrap(io.EOF, "err")}, wantPermErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.RetryEOF(tt.args.err)
			pErr := &backoff.PermanentError{}
			isPErr := errors.As(err, &pErr)

			if isPErr != tt.wantPermErr {
				t.Errorf("RetryEOF() error = %v, wantErr %v", isPErr, tt.wantPermErr)
			}
		})
	}
}
