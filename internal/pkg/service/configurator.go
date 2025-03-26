package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/rs/zerolog/log"

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
	headerAudioSuffix   = "x-tts-audio-suffix"

	defaultVoiceKey = "default"
)

// TTSConfigutaror tts request configuration
type TTSConfigutaror struct {
	defaultOutputFormat api.AudioFormatEnum
	outputMetadata      []string
	availableVoices     map[string]string
	noSSML              bool
}

// NewTTSConfigurator creates the initial request configuration
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

	log.Info().Msgf("Default output format: %s", res.defaultOutputFormat.String())
	res.outputMetadata = cfg.GetStringSlice("output.metadata")
	for _, m := range res.outputMetadata {
		if !strings.Contains(m, "=") {
			return nil, errors.Errorf("metadata must contain '='. Value: '%s'", m)
		}
	}
	log.Info().Msgf("Metadata: %v", res.outputMetadata)

	res.availableVoices, err = initVoices(cfg.GetStringSlice("output.voices"))
	if err != nil {
		return nil, errors.Wrap(err, "can't init voices")
	}
	dVoice, err := getVoice(res.availableVoices, defaultVoiceKey)
	if err != nil {
		return nil, errors.Wrap(err, "no default voice")
	}
	log.Info().Msgf("Voices. Default: %s, all: %v", dVoice, res.availableVoices)
	return res, nil
}

// NewTTSConfiguratorNoSSML creates the initial request configuration with no SSML allowed
func NewTTSConfiguratorNoSSML(cfg *viper.Viper) (*TTSConfigutaror, error) {
	res, err := NewTTSConfigurator(cfg)
	if err != nil {
		return nil, err
	}
	res.noSSML = true
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

// Configure prepares request configuration
func (c *TTSConfigutaror) Configure(ctx context.Context, r *http.Request, inText *api.Input) (*api.TTSRequestConfig, error) {
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

	res.AudioSuffix = getHeader(r, headerAudioSuffix)

	res.Speed, err = getSpeed(inText.Speed)
	if err != nil {
		return nil, err
	}
	res.Voice, err = getVoice(c.availableVoices, inText.Voice)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Info().Str("in", goapp.Sanitize(inText.Voice)).Str("mapped", res.Voice).Msg("voice")
	res.SpeechMarkTypes, err = getSpeechMarkTypes(inText.SpeechMarkTypes)
	if err != nil {
		return nil, err
	}
	res.MaxEdgeSilenceMillis, err = getMaxEdgeSilence(inText.MaxEdgeSilenceMillis)
	if err != nil {
		return nil, err
	}
	log.Ctx(ctx).Info().Int64("edgeSil", res.MaxEdgeSilenceMillis).Any("speechMarks", res.SpeechMarkTypes).Send()
	if inText.Priority < 0 {
		return nil, errors.Errorf("wrong priority (>=0) value: %d", inText.Priority)
	}
	res.Priority = inText.Priority
	res.AllowedMaxLen, err = getMaxLen(getHeader(r, headerMaxTextLen))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(res.Text, "<speak") || inText.TextType == "ssml" {
		if c.noSSML {
			return nil, errors.New("SSML not allowed")
		}
		res.SSMLParts, err = ssml.Parse(strings.NewReader(res.Text),
			&ssml.Text{Voice: res.Voice, Speed: inText.Speed},
			func(s string) (string, error) {
				return getVoice(c.availableVoices, s)
			})
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func getMaxEdgeSilence(value *int64) (int64, error) {
	if value == nil {
		return -1, nil
	}
	if *value < 0 {
		return -1, errors.Errorf("maxEdgeSilenceMillis must be >= 0")
	}
	return *value, nil
}

func getSpeechMarkTypes(s []string) (map[string]bool, error) {
	res := make(map[string]bool)
	for _, v := range s {
		if v != api.SpeechMarkTypeWord {
			return nil, errors.Errorf("Unknown speech mark type '%s'", v)
		}
		res[v] = true
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
	if st == "transcribed" {
		return api.TextTranscribed, nil
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
	return false, errors.Errorf("saveRequest=%t is rejected for this key.", *v)
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
