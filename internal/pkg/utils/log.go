package utils

import "github.com/airenas/go-app/pkg/goapp"

//MaxLogDataSize indicates how many bytes of data to log
var MaxLogDataSize = 100

func logData(st string, data string) {
	goapp.Log.Debugf("%s %s", st, trimString(data, MaxLogDataSize))
}

func trimString(data string, size int) string {
	if len(data) > size {
		return data[0:size] + "..."
	}
	return data
}
