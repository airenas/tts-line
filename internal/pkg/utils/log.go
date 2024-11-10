package utils

import "github.com/airenas/go-app/pkg/goapp"

// MaxLogDataSize indicates how many chars of data to log
var MaxLogDataSize = 100

const (
	_warningMsg        = "Logged data is private and may only be used for error detection!!! Unauthorized use or access may constitute a violation of agreements"
	_topMaxLogDataSize = 10000
)

// LogData logs data to debug
func LogData(msg string, data string, err error) {
	if err != nil {
		// we want to log everything, but using _topMaxLogDataSize to log data in case of error
		// limit very long output, for example audio data may be very long
		goapp.Log.Debug().Err(err).Str("data", goapp.Sanitize(trimString(data, _topMaxLogDataSize))).Str("WARNING", _warningMsg).Msg(msg)
	} else {
		goapp.Log.Debug().Str("data", goapp.Sanitize(trimString(data, MaxLogDataSize))).Msg(msg)
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
