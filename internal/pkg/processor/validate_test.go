package processor

import (
	"strings"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewValidator(200)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewValidator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewValidator(10)
	assert.NotNil(t, err)
	assert.Nil(t, pr)

	_, err = NewValidator(99)
	assert.NotNil(t, err)

	_, err = NewValidator(0)
	assert.NotNil(t, err)
}

func TestInvokeValidator(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewValidator(100)
	assert.NotNil(t, pr)
	d := synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{Text: "olia"}
	err := pr.Process(&d)
	assert.Nil(t, err)
}

func TestInvokeValidator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewValidator(100)
	assert.NotNil(t, pr)
	d := synthesizer.TTSData{}
	d.Input = &api.TTSRequestConfig{Text: strings.Repeat("olia-", 100)}
	err := pr.Process(&d)
	errTL, ok := err.(*utils.ErrTextTooLong)
	if assert.True(t, ok) {
		assert.Equal(t, 100, errTL.Max)
		assert.Equal(t, 500, errTL.Len)
	}
}

func TestInvokeValidator_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewValidator(100)
	err := pr.Process(d)
	assert.Nil(t, err)
}

func Test_getMaxLen(t *testing.T) {
	type args struct {
		def int
		req int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "req", args: args{def: 10, req: 10}, want: 10},
		{name: "req", args: args{def: 10, req: 100}, want: 100},
		{name: "def", args: args{def: 10, req: 0}, want: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMaxLen(tt.args.def, tt.args.req); got != tt.want {
				t.Errorf("getMaxLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	fNone := func(_t *testing.T, _err error) { assert.Nil(_t, _err) }
	fEmpty := func(_t *testing.T, _err error) { assert.Equal(_t, utils.ErrNoInput, _err) }
	type args struct {
		text   string
		maxLen int
	}
	tests := []struct {
		name     string
		args     args
		wantErrF func(*testing.T, error)
	}{
		{name: "empty", args: args{"", 10}, wantErrF: fEmpty},
		{name: "empty trim", args: args{"  ", 10}, wantErrF: fEmpty},
		{name: "too long", args: args{strings.Repeat("a", 11), 10}, wantErrF: testErrTooLong(11, 10)},
		{name: "too long", args: args{strings.Repeat("a", 11), 10}, wantErrF: testErrTooLong(11, 10)},
		{name: "OK", args: args{strings.Repeat("a", 11), 11}, wantErrF: fNone},
		{name: "OK", args: args{strings.Repeat("a", 1000), 500}, wantErrF: testErrTooLong(1000, 500)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErrF(t, validate(getLen(tt.args.text), tt.args.maxLen))
		})
	}
}

func testErrTooLong(l, max int) func(*testing.T, error) {
	return func(t *testing.T, _err error) {
		errTL, ok := _err.(*utils.ErrTextTooLong)
		if assert.True(t, ok) {
			assert.Equal(t, max, errTL.Max)
			assert.Equal(t, l, errTL.Len)
		}
	}
}

func TestInvokeSSMLValidator(t *testing.T) {
	pr, _ := NewSSMLValidator(100)
	assert.NotNil(t, pr)
	d := synthesizer.TTSData{SSMLParts: []*synthesizer.TTSData{{OriginalText: "olia"}}}
	d.Input = &api.TTSRequestConfig{Text: "olia"}
	err := pr.Process(&d)
	assert.Nil(t, err)
}

func TestInvokeSSMLValidator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewSSMLValidator(100)
	assert.NotNil(t, pr)
	d := synthesizer.TTSData{SSMLParts: []*synthesizer.TTSData{{OriginalText: strings.Repeat("olia-", 100)}}}
	d.Input = &api.TTSRequestConfig{Text: "olia"}
	err := pr.Process(&d)
	errTL, ok := err.(*utils.ErrTextTooLong)
	if assert.True(t, ok) {
		assert.Equal(t, 100, errTL.Max)
		assert.Equal(t, 500, errTL.Len)
	}
}

func Test_getSSMLTextLen(t *testing.T) {
	tests := []struct {
		name  string
		parts []*synthesizer.TTSData
		want  int
	}{
		{name: "empty", parts: []*synthesizer.TTSData{}, want: 0},
		{name: "one", parts: []*synthesizer.TTSData{{OriginalText: "olia",
			Cfg: synthesizer.TTSConfig{Type: synthesizer.SSMLText}}}, want: 4},
		{name: "several", parts: []*synthesizer.TTSData{{OriginalText: "olia",
			Cfg: synthesizer.TTSConfig{Type: synthesizer.SSMLText}},
			{OriginalText: "olia2",
				Cfg: synthesizer.TTSConfig{Type: synthesizer.SSMLText}}}, want: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSSMLTextLen(&synthesizer.TTSData{SSMLParts: tt.parts}); got != tt.want {
				t.Errorf("getSSMLTextLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
