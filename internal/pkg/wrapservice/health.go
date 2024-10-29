package wrapservice

import (
	"net/http"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/mschneider82/health"
	"github.com/mschneider82/health/url"
	"github.com/pkg/errors"
)

// NewHealthHandler creates new health handler
func NewHealthHandler(amURL, vocURL string) (http.Handler, error) {
	if amURL == "" {
		return nil, errors.New("No AM URL")
	}
	if vocURL == "" {
		return nil, errors.New("No Vocoder URL")
	}
	timeout := 10 * time.Second
	res := health.NewHandler()
	goapp.Log.Infof("AM live URL: %s", amURL+"/live")
	goapp.Log.Infof("Vocoder live URL: %s", vocURL+"/live")
	res.AddChecker("AM", url.NewCheckerWithTimeout(amURL+"/live", timeout))
	res.AddChecker("Vocoder", url.NewCheckerWithTimeout(vocURL+"/live", timeout))
	return res, nil
}
