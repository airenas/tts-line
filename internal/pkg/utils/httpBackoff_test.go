package utils_test

import (
	"context"
	"fmt"
	"io"
	"syscall"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testHTTPWrap *mocks.HTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	testHTTPWrap = &mocks.HTTPInvokerJSON{}
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
	testHTTPWrap.On("InvokeJSON", mock.Anything, mock.Anything).Return(nil)

	err := pr.InvokeJSON("olia", "")
	assert.Nil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeJSON", 1)
}

func TestInvokeText(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap,
		func() backoff.BackOff { return backoff.NewExponentialBackOff() },
		utils.RetryAll)
	testHTTPWrap.On("InvokeText", mock.Anything, mock.Anything).Return(nil)

	err := pr.InvokeText("olia", "")
	assert.Nil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeText", 1)
}

func TestInvokeText_Retry(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap,
		func() backoff.BackOff {
			return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
		}, utils.RetryAll)
	testHTTPWrap.On("InvokeText", mock.Anything, mock.Anything).Return(errors.New("olia"))

	err := pr.InvokeText("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeText", 4)
}

func TestInvokeRetry(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 3)
	}, utils.RetryAll)
	testHTTPWrap.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))

	err := pr.InvokeJSON("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeJSON", 4)
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
	testHTTPWrap.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))

	err := pr.InvokeJSON("olia", "")
	assert.NotNil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeJSON", 4)
	assert.Equal(t, 4, ic)
	assert.Equal(t, 3, rc)
}

func TestRetry_StopsNonEOF(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 4)
	}, utils.IsRetryable)
	testHTTPWrap.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))
	err := pr.InvokeJSON("olia", "")
	require.NotNil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeJSON", 1)
}

func TestRetry_ContinueEOF(t *testing.T) {
	initTestJSON(t)
	pr, _ := utils.NewHTTPBackoff(testHTTPWrap, func() backoff.BackOff {
		return backoff.WithMaxRetries(&backoff.ZeroBackOff{}, 4)
	}, utils.IsRetryable)
	testHTTPWrap.On("InvokeJSON", mock.Anything, mock.Anything).Return(io.EOF)
	err := pr.InvokeJSON("olia", "")
	require.NotNil(t, err)
	testHTTPWrap.AssertNumberOfCalls(t, "InvokeJSON", 5)
}

func TestIsEOF(t *testing.T) {
	type args struct {
		err error
	}
	sErr := errors.Errorf("olia")
	tests := []struct {
		name      string
		args      args
		wantRetry bool
	}{
		{name: "simple", args: args{err: sErr}, wantRetry: false},
		{name: "EOF", args: args{err: io.EOF}, wantRetry: true},
		{name: "Timeout", args: args{err: context.DeadlineExceeded}, wantRetry: true},
		{name: "Timeout 2", args: args{err: testTmErr{timeout: true}}, wantRetry: true},
		{name: "No Timeout", args: args{err: testTmErr{}}, wantRetry: false},
		{name: "Broken pipe", args: args{err: syscall.EPIPE}, wantRetry: true},
		{name: "Reset by peer", args: args{err: syscall.ECONNRESET}, wantRetry: true},
		{name: "Wrapped EOF", args: args{err: errors.Wrap(io.EOF, "err")}, wantRetry: true},
		{name: "Broken pipe Wrapped", args: args{err: fmt.Errorf("err: %w", syscall.EPIPE)}, wantRetry: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsRetryable(tt.args.err); tt.wantRetry != got {
				t.Errorf("RetryEOF() error = %v, wantErr %v", got, tt.wantRetry)
			}
		})
	}
}

type testTmErr struct {
	timeout, temp bool
}

func (e testTmErr) Error() string {
	return "test err"
}

func (e testTmErr) Timeout() bool {
	return e.timeout
}

func (e testTmErr) Temporary() bool {
	return e.temp
}
