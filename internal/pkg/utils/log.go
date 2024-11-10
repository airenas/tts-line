package utils

import "github.com/airenas/go-app/pkg/goapp"

// MaxLogDataSize indicates how many bytes of data to log
var MaxLogDataSize = 100

const _warningMsg = "Logged data is private and may only be used for error detection. Unauthorized use or access may constitute a violation of agreements."

// LogData logs data to debug
func LogData(msg string, data string, err error) {
	if err != nil {
		goapp.Log.Debug().Err(err).Str("data", goapp.Sanitize(data)).Str("WARNING", _warningMsg).Msg(msg)
	} else {
		goapp.Log.Debug().Str("data", goapp.Sanitize(trimString(data, MaxLogDataSize))).Msg(msg)
	}
}

func trimString(data string, size int) string {
	rn := []rune(data)
	if len(rn) > size {
		return string(rn[:size]) + "..."
	}
	return data
}
