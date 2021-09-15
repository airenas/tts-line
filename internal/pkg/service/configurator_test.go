package service

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestNewTTSConfigurator(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa\n  defaultVoice: aaa"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, "mp3", c.defaultOutputFormat.String())
		assert.Equal(t, []string{"r=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_SeveralMetadata(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa\n   - b=aaa\n  defaultVoice: aaa"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, []string{"r=aaa", "b=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_SeveralVoices(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  defaultVoice: aaa\n  voices:\n   - v1\n   - v2"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, "aaa", c.defaultVoice)
		assert.True(t, c.availableVoices["aaa"])
		assert.True(t, c.availableVoices["v1"])
		assert.True(t, c.availableVoices["v2"])
	}
}

func TestNewTTSConfigurator_Fail(t *testing.T) {
	_, err := NewTTSConfigurator(test.NewConfig(t, ""))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(nil)
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: \n  metadata:\n   - r=aaa\n  defaultVoice: aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: default\n  metadata:\n   - r=aaa\n  defaultVoice: aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: none\n  metadata:\n   - r=aaa\n  defaultVoice: aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp4\n  metadata:\n   - r=aaa\n  defaultVoice: aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - raaa\n  defaultVoice: aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - v1\n   - v2"))
	assert.NotNil(t, err)
}

func TestConfigure_Text(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "olia", res.Text)
		assert.Equal(t, "mp3", res.OutputFormat.String())
		assert.Equal(t, []string{"r=a"}, res.OutputMetadata)
	}
}

func TestConfigure_Format(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4a"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "m4a", res.OutputFormat.String())
	}
}

func TestConfigure_FormatHeader(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerDefaultFormat, "m4a")
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "m4a", res.OutputFormat.String())
	}
}

func TestConfigure_Tags(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerSaveTags, "olia,hola")
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, []string{"olia", "hola"}, res.SaveTags)
	}
}

func TestConfigure_NoTags(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, 0, len(res.SaveTags))
	}
}

func TestConfigure_FailFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4aa"})
	assert.NotNil(t, err)
	req.Header.Add(headerDefaultFormat, "aaa")
	_, err = c.Configure(req, &api.Input{Text: "olia"})
	assert.NotNil(t, err)
}

func TestConfigure_NoneFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "none"})
	assert.Nil(t, err)
	assert.Equal(t, api.AudioNone, res.OutputFormat)
}

func TestConfigure_FailCollect(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	b := true
	_, err := c.Configure(req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.Nil(t, err)
	req.Header.Add(headerCollectData, "never")
	_, err = c.Configure(req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.NotNil(t, err)
}

func TestConfigure_FailTextFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4a", OutputTextFormat: "ooo"})
	assert.NotNil(t, err)
}

func TestConfigure_FailSpeed(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  defaultVoice: mp3"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(req, &api.Input{Text: "olia", Speed: 0.4})
	assert.NotNil(t, err)
}

func TestConfigure_Voice(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - aaa1\n  defaultVoice: aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa", res.Voice)
	}
	res, err = c.Configure(req, &api.Input{Text: "olia", Voice: "aaa"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa", res.Voice)
	}
	res, err = c.Configure(req, &api.Input{Text: "olia", Voice: "aaa1"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa1", res.Voice)
	}
	_, err = c.Configure(req, &api.Input{Text: "olia", Voice: "aaa2"})
	assert.NotNil(t, err)
}

func TestOutputTextFormat(t *testing.T) {
	tests := []struct {
		in    string
		res   api.TextFormatEnum
		isErr bool
	}{
		{in: "", res: api.TextNone, isErr: false},
		{in: "none", res: api.TextNone, isErr: false},
		{in: "normalized", res: api.TextNormalized, isErr: false},
		{in: "accented", res: api.TextAccented, isErr: false},
		{in: "olia", res: api.TextNone, isErr: true},
	}

	for _, tc := range tests {
		v, err := getOutputTextFormat(tc.in)
		assert.Equal(t, tc.res, v)
		assert.Equal(t, tc.isErr, err != nil)
	}
}

func TestOutputAudioFormat(t *testing.T) {
	tests := []struct {
		in    string
		res   api.AudioFormatEnum
		isErr bool
	}{
		{in: "", res: api.AudioDefault, isErr: false},
		{in: "none", res: api.AudioNone, isErr: false},
		{in: "mp3", res: api.AudioMP3, isErr: false},
		{in: "m4a", res: api.AudioM4A, isErr: false},
		{in: "olia", res: api.AudioNone, isErr: true},
	}

	for _, tc := range tests {
		v, err := getOutputAudioFormat(tc.in)
		assert.Equal(t, tc.res, v)
		assert.Equal(t, tc.isErr, err != nil)
	}
}

func TestAllowCollect(t *testing.T) {
	bt := true
	bf := false
	tests := []struct {
		v     *bool
		h     string
		res   bool
		isErr bool
	}{
		{v: nil, h: "", res: false, isErr: false},
		{v: &bf, h: "", res: false, isErr: false},
		{v: &bt, h: "", res: true, isErr: false},
		{v: nil, h: "never", res: false, isErr: false},
		{v: nil, h: "always", res: true, isErr: false},
		{v: &bt, h: "always", res: true, isErr: false},
		{v: &bf, h: "never", res: false, isErr: false},
		{v: &bf, h: "always", res: false, isErr: true},
		{v: &bt, h: "never", res: false, isErr: true},
	}

	for i, tc := range tests {
		v, err := getAllowCollect(tc.v, tc.h)
		assert.Equal(t, tc.res, v, "fail %d", i)
		assert.Equal(t, tc.isErr, err != nil, "fail %d", i)
	}
}

func TestSpeedValue(t *testing.T) {
	tests := []struct {
		v     float32
		e     float32
		isErr bool
	}{
		{v: 0, e: 0, isErr: false},
		{v: 0.1, e: 0, isErr: true},
		{v: -20, e: 0, isErr: true},
		{v: 0.4999, e: 0, isErr: true},
		{v: 0.5, e: 0.5, isErr: false},
		{v: 2, e: 2, isErr: false},
		{v: 1, e: 1, isErr: false},
	}

	for i, tc := range tests {
		v, err := getSpeed(tc.v)
		assert.InDelta(t, tc.e, v, 0.0001, "fail %d", i)
		assert.Equal(t, tc.isErr, err != nil, "fail %d  - %v", i, err)
	}
}

func TestVoices(t *testing.T) {
	tests := []struct {
		v     string
		av    []string
		e     string
		isErr bool
	}{
		{v: "aaa", av: []string{"aaa", "aaa1"}, e: "aaa", isErr: false},
		{v: "aaa2", av: []string{"aaa", "aaa1"}, e: "", isErr: true},
		{v: "aaa2", av: []string{"aaa", "aaa2"}, e: "aaa2", isErr: false},
		{v: "", av: []string{"aaa", "aaa1"}, e: "aaa", isErr: false},
		{v: "", av: []string{"aaa"}, e: "aaa", isErr: false},
		{v: "aaa", av: []string{"aaa"}, e: "aaa", isErr: false},
		{v: "aaa2", av: []string{"aaa"}, e: "", isErr: true},
	}

	for i, tc := range tests {
		c := TTSConfigutaror{}
		c.defaultVoice, c.availableVoices, _ = initVoices(tc.av[0], tc.av[1:])
		v, err := c.getVoice(tc.v)
		assert.Equal(t, tc.e, v, "fail %d", i)
		assert.Equal(t, tc.isErr, err != nil, "fail %d  - %v", i, err)
	}
}
