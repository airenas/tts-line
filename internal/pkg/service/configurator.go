package service

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/spf13/viper"

	"github.com/pkg/errors"
)

const (
	headerDefaultFormat = "x-tts-default-output-format"
	headerCollectData   = "x-tts-collect-data"
	headerSaveTags      = "x-tts-save-tags"
	headerMaxTextLen    = "x-tts-max-text-len"

	defaultVoiceKey = "default"
)

//TTSConfigutaror tts request configuration
type TTSConfigutaror struct {
	defaultOutputFormat api.AudioFormatEnum
	outputMetadata      []string
	availableVoices     map[string]string
}

//NewTTSConfigurator creates the initial request configuration
func NewTTSConfigurator(cfg *viper.Viper) (*TTSConfigutaror, error) {
	if cfg == nil {
		return nil, errors.New("no request config 'options'")
	}
	res := &TTSConfigutaror{}
	var err error
	res.defaultOutputFormat, err = getOutputAudioFormat(cfg.GetString("output.defaultFormat"))
	if err != nil {
		return nil, errors.Wrap(err, "can't init format")
	}
	if res.defaultOutputFormat == api.AudioNone || res.defaultOutputFormat == api.AudioDefault {
		return nil, errors.New("no output.defaultFormat configured")
	}

	goapp.Log.Infof("Default output format: %s", res.defaultOutputFormat.String())
	res.outputMetadata = cfg.GetStringSlice("output.metadata")
	for _, m := range res.outputMetadata {
		if !strings.Contains(m, "=") {
			return nil, errors.Errorf("metadata must contain '='. Value: '%s'", m)
		}
	}
	goapp.Log.Infof("Metadata: %v", res.outputMetadata)

	res.availableVoices, err = initVoices(cfg.GetStringSlice("output.voices"))
	if err != nil {
		return nil, errors.Wrap(err, "can't init voices")
	}
	dVoice, err := getVoice(res.availableVoices, defaultVoiceKey)
	if err != nil {
		return nil, errors.Wrap(err, "no default voice")
	}
	goapp.Log.Infof("Voices. Default: %s, all: %v", dVoice, res.availableVoices)
	return res, nil
}

func getVoice(voices map[string]string, voiceKey string) (string, error) {
	key := voiceKey
	if key == "" {
		key = defaultVoiceKey
	}
	res, ok := voices[key]
	if !ok {
		return "", errors.Errorf("unknown voice '%s'", key)
	}
	//try go deeper
	// allow mapping default:astra.latest, astra.latest:astra.v02
	for i := 0; i < 3; i++ {
		resN, ok := voices[res]
		if !ok {
			break
		}
		res = resN
	}
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
	res.Voice, err = getVoice(c.availableVoices, inText.Voice)
	if err != nil {
		return nil, err
	}
	goapp.Log.Infof("Voice '%s' -> '%s'", inText.Voice, res.Voice)
	if inText.Priority < 0 {
		return nil, errors.Errorf("wrong priority (>=0) value: %d", inText.Priority)
	}
	res.Priority = inText.Priority
	res.AllowedMaxLen, err = getMaxLen(getHeader(r, headerMaxTextLen))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(res.Text, "<speak") || inText.TextType == "ssml" {
		res.SSMLParts, err = ssml.Parse(strings.NewReader(res.Text),
			&ssml.Text{Voice: res.Voice, Speed: inText.Speed})
		if err != nil {
			return nil, err
		}
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

func initVoices(all []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, s := range all {
		s = strings.TrimSpace(s)
		strs := strings.Split(s, ":")
		if len(strs) != 2 {
			return nil, errors.Errorf("wrong voice value '%s'", s)
		}
		strs[0], strs[1] = strings.TrimSpace(strs[0]), strings.TrimSpace(strs[1])
		if strs[0] == "" || strs[1] == "" {
			return nil, errors.Errorf("wrong voice value '%s'", s)
		}
		res[strs[0]] = strs[1]
	}
	// add values as keys if not exists
	for _, v := range res {
		if _, ok := res[v]; !ok {
			res[v] = v
		}
	}
	return res, nil
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
			return 0, errors.Errorf("speed value (%.2f) must be in [0.5,2].", v)
		}
	}
	return v, nil
}

func getMaxLen(s string) (int, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	res, err := strconv.ParseInt(s, 10, 32)
	return int(res), err
}

func defaultS(s, s1 string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return s1
}
