package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"syscall"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// HTTPInvokerJSON http POST invoker with JSON in, out params
type HTTPInvokerJSON interface {
	InvokeJSONU(context.Context, string, interface{}, interface{}) error
	InvokeJSON(context.Context, interface{}, interface{}) error
	InvokeText(context.Context, string, interface{}) error
}

// HTTPBackoff http call with backoff
type HTTPBackoff struct {
	HTTPClient          HTTPInvokerJSON
	backoffF            func() backoff.BackOff
	retryF              func(error) bool
	InvokeIndicatorFunc func(interface{})
	RetryIndicatorFunc  func(interface{})
}

// NewHTTPBackoff creates new wrapper with backoff
// errF must return same error or wrap to backoff.PermanentError to stop backoff
func NewHTTPBackoff(realWrapper HTTPInvokerJSON, backoffF func() backoff.BackOff, retryF func(error) bool) (*HTTPBackoff, error) {
	return &HTTPBackoff{HTTPClient: realWrapper, backoffF: backoffF, retryF: retryF}, nil
}

// InvokeJSON makes http call with json
func (hw *HTTPBackoff) InvokeJSON(ctx context.Context, dataIn interface{}, dataOut interface{}) error {
	return hw.invoke(ctx, func(_ctx context.Context) error {
		return hw.HTTPClient.InvokeJSON(_ctx, dataIn, dataOut)
	}, dataIn)
}

// InvokeText makes http call with text input
func (hw *HTTPBackoff) InvokeText(ctx context.Context, dataIn string, dataOut interface{}) error {
	return hw.invoke(ctx, func(_ctx context.Context) error {
		return hw.HTTPClient.InvokeText(_ctx, dataIn, dataOut)
	}, dataIn)
}

// InvokeJSONU makes call to URL wits JSON
func (hw *HTTPBackoff) InvokeJSONU(ctx context.Context, URL string, dataIn interface{}, dataOut interface{}) error {
	return hw.invoke(ctx, func(_ctx context.Context) error {
		return hw.HTTPClient.InvokeJSONU(_ctx, URL, dataIn, dataOut)
	}, dataIn)
}

func (hw *HTTPBackoff) invoke(ctx context.Context, f func(context.Context) error, dataIn interface{}) error {
	ctx, span := StartSpan(ctx, "HTTPBackoff.invoke")
	defer span.End()

	failC := 0
	op := func() error {
		if hw.InvokeIndicatorFunc != nil {
			hw.InvokeIndicatorFunc(dataIn)
		}
		if failC > 0 && hw.RetryIndicatorFunc != nil {
			hw.RetryIndicatorFunc(dataIn)
		}
		err := f(ctx)
		if err != nil {
			failC++
			if !hw.retryF(err) {
				return backoff.Permanent(err)
			}
			select {
			case <-ctx.Done(): // do not retry if context is done
				errCtx := ctx.Err()
				if errCtx != nil && err != errCtx {
					err = fmt.Errorf("%w: %w", errCtx, err)
				}
				return backoff.Permanent(err)
			default:
			}
			log.Ctx(ctx).Warn().Err(errors.Wrapf(err, "failed %d time(s)", failC)).Send()
			return err
		}
		return nil
	}
	err := backoff.Retry(op, hw.backoffF())
	if err == nil && failC > 0 {
		log.Ctx(ctx).Info().Msgf("Success after retrying %d time(s)", failC)
	}
	return err
}

// RetryAll - retries all errors
func RetryAll(err error) bool {
	return err != nil
}

// IsRetryable - check if error is retryable io.EOF or timeout
func IsRetryable(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) ||
		isTimeout(err)
}

func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && (e.Timeout())
}

// Info returns info about wrapper
func (hw *HTTPBackoff) Info() string {
	return fmt.Sprintf("HTTPBackoff(%s)", RetrieveInfo(hw.HTTPClient))
}
