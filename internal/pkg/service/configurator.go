package service

import (
	"net/http"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/spf13/viper"

	"github.com/pkg/errors"
)

const (
	headerDefaultFormat = "x-tts-default-output-format"
	headerCollectData   = "x-tts-collect-data"
	headerSaveTags      = "x-tts-save-tags"
)

//TTSConfigutaror tts request configuration
type TTSConfigutaror struct {
	defaultOutputFormat api.AudioFormatEnum
	outputMetadata      []string
}

//NewTTSConfigurator creates the initial request configuration
func NewTTSConfigurator(cfg *viper.Viper) (*TTSConfigutaror, error) {
	if cfg == nil {
		return nil, errors.New("No request config 'options'")
	}
	res := &TTSConfigutaror{}
	var err error
	res.defaultOutputFormat, err = getOutputAudioFormat(cfg.GetString("output.defaultFormat"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't init format")
	}
	if res.defaultOutputFormat == api.AudioNone || res.defaultOutputFormat == api.AudioDefault {
		return nil, errors.New("No output.defaultFormat configured")
	}

	goapp.Log.Infof("Default output format: %s", res.defaultOutputFormat.String())
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
	var err error
	res.OutputFormat, err = getOutputAudioFormat(defaultS(inText.OutputFormat, getHeader(r, headerDefaultFormat)))
	if err != nil {
		return nil, err
	}
	if res.OutputFormat == api.AudioDefault {
		res.OutputFormat = c.defaultOutputFormat
	}

	res.OutputMetadata = c.outputMetadata
	res.OutputTextFormat, err = getOutputTextFormat(inText.OutputTextFormat)
	if err != nil {
		return nil, err
	}
	res.AllowCollectData, err = getAllowCollect(inText.AllowCollectData, getHeader(r, headerCollectData))
	if err != nil {
		return nil, err
	}
	res.SaveTags = getSaveTags(getHeader(r, headerSaveTags))

	res.Speed, err = getSpeed(inText.Speed)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getOutputTextFormat(s string) (api.TextFormatEnum, error) {
	st := strings.TrimSpace(s)
	if st == "" || st == "none" {
		return api.TextNone, nil
	}
	if st == "normalized" {
		return api.TextNormalized, nil
	}
	if st == "accented" {
		return api.TextAccented, nil
	}
	return api.TextNone, errors.New("Unknown text format " + s)
}

func getAllowCollect(v *bool, s string) (bool, error) {
	st := strings.TrimSpace(strings.ToLower(s))
	if st == "" || st == "request" {
		return v != nil && *v, nil
	}
	if v == nil {
		return st == "always", nil
	}
	if st == "always" && *v {
		return true, nil
	}
	if st == "never" && !*v {
		return false, nil
	}
	return false, errors.Errorf("AllowCollectData=%t is rejected for this key.", *v)
}

func getOutputAudioFormat(s string) (api.AudioFormatEnum, error) {
	st := strings.TrimSpace(s)
	if st == "" {
		return api.AudioDefault, nil
	}
	if st == "mp3" {
		return api.AudioMP3, nil
	}
	if st == "m4a" {
		return api.AudioM4A, nil
	}
	if st == "none" {
		return api.AudioNone, nil
	}
	return api.AudioNone, errors.New("Unknown audio format " + s)
}

func getHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func getSaveTags(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(v), ",")
}

func getSpeed(v float32) (float32, error) {
	if !utils.FloatEquals(v, 0) {
		if v < 0.5 || v > 2.0 {
			return 0, errors.Errorf("speed value (%.3f) must be in [0.5,2]. ", v)
		}
	}
	return v, nil
}

func defaultS(s, s1 string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return s1
}
