package utils

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func checkURL(urlStr string) (string, error) {
	if strings.TrimSpace(urlStr) == "" {
		return "", errors.New("No url")
	}
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return "", errors.New("Can't parse url " + urlStr)
	}
	return urlRes.String(), nil
}

const eps float32 = 0.00000001
func FloatEquals(a, b float32) bool {
	return (a - b) < eps && (b - a) < eps
}
