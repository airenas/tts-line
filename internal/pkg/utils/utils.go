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

// FloatEquals compares two floats
func FloatEquals(a, b float32) bool {
	return (a-b) < eps && (b-a) < eps
}

// RetrieveInfo invokes Info() if interface has one
// else returns ""
func RetrieveInfo(pr interface{}) string {
	pri, ok := pr.(interface {
		Info() string
	})
	if ok {
		return pri.Info()
	}
	return ""
}
