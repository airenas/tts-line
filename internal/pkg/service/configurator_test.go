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
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, "mp3", c.defaultOutputFormat.String())
		assert.Equal(t, []string{"r=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_SeveralMetadata(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa\n   - b=aaa"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, []string{"r=aaa", "b=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_Fail(t *testing.T) {
	_, err := NewTTSConfigurator(test.NewConfig(t, ""))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(nil)
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: \n  metadata:\n   - r=aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp4\n  metadata:\n   - r=aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - raaa"))
	assert.NotNil(t, err)
}

func TestConfigure_Text(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a"))
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
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4a"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "m4a", res.OutputFormat.String())
	}
}

func TestConfigure_FormatHeader(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerDefaultFormat, "m4a")
	res, err := c.Configure(req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "m4a", res.OutputFormat.String())
	}
}

func TestConfigure_FailFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4aa"})
	assert.NotNil(t, err)
	req.Header.Add(headerDefaultFormat, "aaa")
	_, err = c.Configure(req, &api.Input{Text: "olia"})
	assert.NotNil(t, err)
}

func TestConfigure_FailCollect(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	b := true
	_, err := c.Configure(req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.Nil(t, err)
	req.Header.Add(headerCollectData, "never")
	_, err = c.Configure(req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.NotNil(t, err)
}

func TestConfigure_FailTextFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(req, &api.Input{Text: "olia", OutputFormat: "m4a", OutputTextFormat: "ooo"})
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
		{in: "", res: api.AudioNone, isErr: false},
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
