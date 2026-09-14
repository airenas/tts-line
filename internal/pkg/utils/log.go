package utils

import (
	"context"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// MaxLogDataSize indicates how many chars of data to log
var MaxLogDataSize = 100

const (
	_warningMsg        = "Logged data is private and may only be used for error detection!!! Unauthorized use or access may constitute a violation of agreements"
	_topMaxLogDataSize = 10000
)

// LogData logs data to debug
func LogData(ctx context.Context, msg string, data string, err error) {
	if err != nil {
		// we want to log everything, but using _topMaxLogDataSize to log data in case of error
		// limit very long output, for example audio data may be very long
		log.Ctx(ctx).Debug().Err(err).Str("data", goapp.Sanitize(trimString(data, _topMaxLogDataSize))).Str("WARNING", _warningMsg).Msg(msg)
	} else {
		log.Ctx(ctx).Debug().Str("data", goapp.Sanitize(trimString(data, MaxLogDataSize))).Msg(msg)
	}
}

func trimString(data string, size int) string {
	// fast check without rune conversion
	if len(data) < size {
		return data
	}

	rn := []rune(data)
	if len(rn) > size {
		return string(rn[:size]) + "..."
	}
	return data
}

//nolint:zerologlint // The returned event is completed by the caller.
func prepareLog(c echo.Context, v middleware.RequestLoggerValues) *zerolog.Event {
	if v.Status == 404 || v.Status == 405 {
		return log.Ctx(c.Request().Context()).Info().Err(v.Error)
	}
	if v.Status >= 400 || v.Error != nil {
		return log.Ctx(c.Request().Context()).Error().Err(v.Error)
	}
	if v.URIPath == "/live" || v.URIPath == "/metrics" {
		return log.Ctx(c.Request().Context()).Trace()
	}

	return log.Ctx(c.Request().Context()).Info()
}

func EchoLogMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogURIPath:       true,
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogRemoteIP:      true,
		LogUserAgent:     true,
		LogResponseSize:  true,
		LogContentLength: true,
		LogHost:          true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			prepareLog(c, v).
				Str("uri", v.URI).
				Str("remote_ip", v.RemoteIP).
				Str("user_agent", v.UserAgent).
				Int("status", v.Status).
				Str("latency_human", v.Latency.String()).
				Str("bytes_in", v.ContentLength).
				Int64("bytes_out", v.ResponseSize).
				Str("host", v.Host).
				Send()
			return nil
		},
	})
}
