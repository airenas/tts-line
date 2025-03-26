package transcription

import (
	"strings"
)

type Data struct {
	Word          string
	Sylls         string
	Transcription string
}

func Parse(str string) *Data {
	var res Data
	res.Transcription = strings.ReplaceAll(str, "-", "")
	res.Sylls = TrimAccent(str)
	res.Sylls = strings.ToLower(res.Sylls)
	res.Word = strings.ReplaceAll(res.Sylls, "-", "")
	return &res
}

func TrimAccent(s string) string {
	r := strings.ReplaceAll(s, "3", "")
	r = strings.ReplaceAll(r, "4", "")
	r = strings.ReplaceAll(r, "9", "")
	return r
}
