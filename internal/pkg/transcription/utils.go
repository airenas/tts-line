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
	strTmp := strings.ReplaceAll(str, "Q", "O")

	res.Transcription = strings.ReplaceAll(strTmp, "-", "")
	res.Sylls = TrimAccent(strTmp)
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
