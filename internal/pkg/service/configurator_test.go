package service

import (
	"context"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTTSConfigurator(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa\n  voices:\n   - default:astra"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, api.AudioMP3.String(), c.defaultOutputFormat.String())
		assert.Equal(t, []string{"r=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_SeveralMetadata(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=aaa\n   - b=aaa\n  voices:\n   - default:astra"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, []string{"r=aaa", "b=aaa"}, c.outputMetadata)
	}
}

func TestNewTTSConfigurator_SeveralVoices(t *testing.T) {
	c, err := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:v1\n   - v1:v1\n   - v2:v2"))
	assert.Nil(t, err)
	if assert.NotNil(t, c) {
		assert.Equal(t, "v1", c.availableVoices["default"])
		assert.Equal(t, "v1", c.availableVoices["v1"])
		assert.Equal(t, "v2", c.availableVoices["v2"])
	}
}

func TestNewTTSConfigurator_Fail(t *testing.T) {
	_, err := NewTTSConfigurator(test.NewConfig(t, ""))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(nil)
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: \n  metadata:\n   - r=aaa\n  voices:\n   - default:aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: default\n  metadata:\n   - r=aaa\n  voices:\n   - default:aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: none\n  metadata:\n   - r=aaa\n  voices:\n   - default:aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp4\n  metadata:\n   - r=aaa\n  voices:\n   - default:aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - raaa\n  voices:\n   - default:aaa"))
	assert.NotNil(t, err)
	_, err = NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - v1\n   - v2"))
	assert.NotNil(t, err)
}

func TestConfigure_Text(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "olia", res.Text)
		assert.Equal(t, api.AudioMP3, res.OutputFormat)
		assert.Equal(t, []string{"r=a"}, res.OutputMetadata)
		assert.Equal(t, 0, res.AllowedMaxLen)
	}
}

func TestConfigure_Format(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", OutputFormat: "m4a"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, api.AudioM4A, res.OutputFormat)
	}
}

func TestConfigure_FormatHeader(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerDefaultFormat, "m4a")
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, api.AudioM4A, res.OutputFormat)
	}
}

func TestConfigure_FormatHeaderWAV(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerDefaultFormat, "wav")
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	require.NotNil(t, res)
	assert.Equal(t, api.AudioWAV, res.OutputFormat)
}

func TestConfigure_MaxTextLen(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerMaxTextLen, "102")
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, 102, res.AllowedMaxLen)
	}
}

func TestConfigure_MaxTextLenFail(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerMaxTextLen, "a102")
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.NotNil(t, err)
}

func TestConfigure_Tags(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerSaveTags, "olia,hola")
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, []string{"olia", "hola"}, res.SaveTags)
	}
}

func TestConfigure_NoTags(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, 0, len(res.SaveTags))
	}
}

func TestConfigure_FailFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", OutputFormat: "m4aa"})
	assert.NotNil(t, err)
	req.Header.Add(headerDefaultFormat, "aaa")
	_, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.NotNil(t, err)
}

func TestConfigure_NoneFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  metadata:\n   - r=a\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", OutputFormat: "none"})
	assert.Nil(t, err)
	assert.Equal(t, api.AudioNone, res.OutputFormat)
}

func TestConfigure_FailCollect(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	b := true
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.Nil(t, err)
	req.Header.Add(headerCollectData, "never")
	_, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", AllowCollectData: &b})
	assert.NotNil(t, err)
}

func TestConfigure_FailTextFormat(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", OutputFormat: "m4a", OutputTextFormat: "ooo"})
	assert.NotNil(t, err)
}

func TestConfigure_FailSpeed(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", Speed: 0.4})
	assert.NotNil(t, err)
}

func TestConfigure_Voice(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - aaa1:aaa2\n   - aaa:aaa\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa", res.Voice)
	}
	res, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", Voice: "aaa"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa", res.Voice)
	}
	res, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", Voice: "aaa1"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa2", res.Voice)
	}
	res, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", Voice: "aaa2"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "aaa2", res.Voice)
	}
	_, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", Voice: "aaa3"})
	assert.NotNil(t, err)
}

func TestConfigure_Priority(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	require.NotNil(t, c)
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", Priority: 10})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, 10, res.Priority)
	}
	res, err = c.Configure(context.TODO(), req, &api.Input{Text: "olia", Priority: -1})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestConfigure_AudioSuffix(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	req.Header.Add(headerAudioSuffix, "olia.wav")
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, "olia.wav", res.AudioSuffix)
	}
}

func TestConfigure_SSML(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	res, err := c.Configure(context.TODO(), req, &api.Input{Text: "<speak>olia</speak>"})
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, 1, len(res.SSMLParts))
	}
}

func TestConfigure_SSML_Fail_NoSSML(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "olia", TextType: "ssml"})
	assert.NotNil(t, err)
}

func TestConfigure_SSML_Fail_BadSSML(t *testing.T) {
	c, _ := NewTTSConfigurator(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "<speak><k></k></speak>", TextType: "ssml"})
	assert.NotNil(t, err)
}

func TestConfigure_SSML_NoAllowed(t *testing.T) {
	c, _ := NewTTSConfiguratorNoSSML(test.NewConfig(t, "output:\n  defaultFormat: mp3\n  voices:\n   - default:aaa"))
	req := httptest.NewRequest("POST", "/synthesize", strings.NewReader("text"))
	_, err := c.Configure(context.TODO(), req, &api.Input{Text: "<speak>olia</speak>", TextType: "ssml"})
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
		{in: "transcribed", res: api.TextTranscribed, isErr: false},
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

	for i, tt := range tests {
		t.Run(tt.h, func(t *testing.T) {
			v, err := getAllowCollect(tt.v, tt.h)
			assert.Equal(t, tt.res, v, "fail %d", i)
			assert.Equal(t, tt.isErr, err != nil, "fail %d", i)
		})
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

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%f", tt.v), func(t *testing.T) {
			v, err := getSpeed(tt.v)
			assert.InDelta(t, tt.e, v, 0.0001, "fail %d", i)
			assert.Equal(t, tt.isErr, err != nil, "fail %d  - %v", i, err)
		})
	}
}

func TestVoices(t *testing.T) {
	tests := []struct {
		v     string
		av    map[string]string
		e     string
		isErr bool
	}{
		{v: "aaa", av: map[string]string{"aaa": "aaa", "aaa1": "aaa1"}, e: "aaa", isErr: false},
		{v: "aaa2", av: map[string]string{"aaa": "aaa", "aaa1": "aaa1"}, e: "", isErr: true},
		{v: "aaa2", av: map[string]string{"aaa": "aaa", "aaa2": "aaa2"}, e: "aaa2", isErr: false},
		{v: "", av: map[string]string{"default": "aaa", "aaa2": "aaa2"}, e: "aaa", isErr: false},
		{v: "", av: map[string]string{"default": "aaa", "aaa": "aaa2"}, e: "aaa2", isErr: false},
		{v: "in", av: map[string]string{"in": "aaa.latest", "aaa.latest": "aaa2", "aaa2": "aaa"}, e: "aaa", isErr: false},
		{v: "a", av: map[string]string{"aa": "aa", "aaa2": "a"}, e: "", isErr: true},
		{v: "rec", av: map[string]string{"rec": "rec1", "rec1": "rec2", "rec2": "rec"}, e: "rec1", isErr: false},
	}

	for i, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			v, err := getVoice(tt.av, tt.v)
			assert.Equal(t, tt.e, v, "fail %d", i)
			assert.Equal(t, tt.isErr, err != nil, "fail %d  - %v", i, err)
		})
	}
}

func Test_getMaxLen(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    int
		wantErr bool
	}{
		{name: "OK", args: "10", want: 10, wantErr: false},
		{name: "OK", args: "10234", want: 10234, wantErr: false},
		{name: "Fail", args: "a10", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMaxLen(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMaxLen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getMaxLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSpeechMarkTypes(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{name: "OK", args: args{s: []string{"word"}}, want: map[string]bool{"word": true}, wantErr: false},
		{name: "Fail", args: args{s: []string{"word1"}}, want: nil, wantErr: true},
		{name: "Empty", args: args{s: nil}, want: map[string]bool{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSpeechMarkTypes(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSpeechMarkTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSpeechMarkTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMaxEdgeSilence(t *testing.T) {
	type args struct {
		value *int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{name: "OK", args: args{value: nil}, want: -1, wantErr: false},
		{name: "OK", args: args{value: &[]int64{10}[0]}, want: 10, wantErr: false},
		{name: "Fail", args: args{value: &[]int64{-10}[0]}, want: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMaxEdgeSilence(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMaxEdgeSilence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getMaxEdgeSilence() = %v, want %v", got, tt.want)
			}
		})
	}
}
