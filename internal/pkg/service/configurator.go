package service

import (
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/spf13/viper"

	"github.com/pkg/errors"
)

const (
	headerDefaultFormat = "x-tts-default-output-format"
)

//TTSConfigutaror tts request configuration
type TTSConfigutaror struct {
	defaultOutputFormat string
	outputMetadata      []string
}

//NewTTSConfigurator creates the initial request configuration
func NewTTSConfigurator(cfg *viper.Viper) (*TTSConfigutaror, error) {
	if cfg == nil {
		return nil, errors.New("No request config 'options'")
	}
	res := &TTSConfigutaror{}
	res.defaultOutputFormat = cfg.GetString("output.defaultFormat")
	if res.defaultOutputFormat == "" {
		return nil, errors.New("No output.defaultFormat configured")
	}
	goapp.Log.Infof("Default output format: %s", res.defaultOutputFormat)
	res.outputMetadata = cfg.GetStringSlice("output.metadata")
	for _, m := range res.outputMetadata {
		if !strings.Contains(m, "=") {
			return nil, errors.Errorf("Metadata must contain '='. Value: '%s'", m)
		}
	}
	goapp.Log.Infof("Metadata: %v", res.outputMetadata)
	return res, nil
}

//Configure prepares request configuration
func (c *TTSConfigutaror) Configure(r *http.Request, inText *api.Input) (*api.TTSRequestConfig, error) {
	res := &api.TTSRequestConfig{}
	res.Text = inText.Text
	res.OutputFormat = inText.OutputFormat
	if res.OutputFormat == "" {
		res.OutputFormat = getHeader(r, headerDefaultFormat)
	}
	if res.OutputFormat == "" {
		res.OutputFormat = c.defaultOutputFormat
	}
	if res.OutputFormat != "mp3" && res.OutputFormat != "m4a" {
		return nil, errors.Errorf("Unsuported format '%s'", res.OutputFormat)
	}
	res.OutputMetadata = c.outputMetadata
	return res, nil
}

func getHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}
