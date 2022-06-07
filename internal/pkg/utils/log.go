package utils

import "github.com/airenas/go-app/pkg/goapp"

//MaxLogDataSize indicates how many bytes of data to log
var MaxLogDataSize = 100

//LogData logs data to debug
func LogData(st string, data string) {
	goapp.Log.Debugf("%s %s", st, goapp.Sanitize(trimString(data, MaxLogDataSize)))
}

func trimString(data string, size int) string {
	if len(data) > size {
		return data[:size] + "..."
	}
	return data
}
