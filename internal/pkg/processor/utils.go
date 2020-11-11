package processor

import (
	"net/url"

	"github.com/pkg/errors"
)

func checkURL(urlStr string) (string, error) {
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return "", errors.New("Can't parse url " + urlStr)
	}
	return urlRes.String(), nil
}
