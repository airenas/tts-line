package utils

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

//ErrNoRecord indicates no record found error
var ErrNoRecord = errors.New("no record found")

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
