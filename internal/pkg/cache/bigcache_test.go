package cache

import (
	"strings"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	synthesizerMock *mocks.Synthesizer
)

func initTest(t *testing.T) {
	synthesizerMock = &mocks.Synthesizer{}
}

func TestNewCacher(t *testing.T) {
	initTest(t)
	c, err := NewCacher(synthesizerMock, newTestConfig(""))
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Nil(t, c.cache)
}

func TestNewCacher_Fails(t *testing.T) {
	initTest(t)
	_, err := NewCacher(nil, newTestConfig(""))
	assert.NotNil(t, err)
}

func TestNewCacherInit(t *testing.T) {
	initTest(t)
	c, err := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.cache)
}

func TestCleanDuartion(t *testing.T) {
	assert.Equal(t, 5*time.Minute, getCleanDuration(0))
	assert.Equal(t, 2*time.Second, getCleanDuration(2*time.Second))
}

func TestWork(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(&api.Result{AudioAsString: "wav"}, nil)

	res, err := c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	assert.Equal(t, "wav", res.AudioAsString)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	res, err = c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	assert.Equal(t, "wav", res.AudioAsString)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	res, err = c.Work(newtestInput("olia2"))
	assert.Nil(t, err)
	assert.NotNil(t, res)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
}

func TestWork_Failure(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(nil, errors.New("haha"))

	_, err := c.Work(newtestInput("olia"))
	assert.NotNil(t, err)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	_, err = c.Work(newtestInput("olia"))
	assert.NotNil(t, err)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
}

func TestWork_NoCache(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 0s"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(&api.Result{AudioAsString: "wav"}, nil)

	_, err := c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	_, err = c.Work(newtestInput("olia"))
	assert.Nil(t, err)
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
}

func TestWork_Key(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(&api.Result{AudioAsString: "wav"}, nil)

	_, _ = c.Work(newtestInput("olia"))
	_, _ = c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3})
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
	_, _ = c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3})
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
	_, _ = c.Work(&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioM4A})
	synthesizerMock.AssertNumberOfCalls(t, "Work", 3)
}

func Test_MaxMB(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s\nmaxMB: 1"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(&api.Result{AudioAsString: strOfSize(1024 * 1024 / 64)}, nil) // 64 shards in cache hardcoded

	_, _ = c.Work(newtestInput("olia"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	_, _ = c.Work(newtestInput("olia"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2) //expected not to add
	_, _ = c.Work(newtestInput("olia"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 3)
}

func Test_MaxTextLen(t *testing.T) {
	initTest(t)
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s\nmaxTextLen: 10"))
	assert.NotNil(t, c)
	synthesizerMock.On("Work", mock.Anything).Return(&api.Result{AudioAsString: "wav"}, nil) // 64 shards in cache hardcoded

	_, _ = c.Work(newtestInput("0123456789"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)
	_, _ = c.Work(newtestInput("0123456789"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 1)

	_, _ = c.Work(newtestInput("01234567891"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 2)
	_, _ = c.Work(newtestInput("01234567891"))
	synthesizerMock.AssertNumberOfCalls(t, "Work", 3)
}

func newTestConfig(yaml string) *viper.Viper {
	res := viper.New()
	res.SetConfigType("yaml")
	_ = res.ReadConfig(strings.NewReader(yaml))
	return res
}

func newtestInput(txt string) *api.TTSRequestConfig {
	return &api.TTSRequestConfig{Text: txt}
}

func strOfSize(s int) string {
	return string(make([]byte, s))
}

func TestBigCacher_isOK(t *testing.T) {
	c, _ := NewCacher(synthesizerMock, newTestConfig("duration: 10s\nmaxTextLen: 10"))
	type args struct {
		inp *api.TTSRequestConfig
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"OK", args{&api.TTSRequestConfig{Text: "aaa", OutputTextFormat: api.TextNone}}, true},
		{"Accented", args{&api.TTSRequestConfig{Text: "aaa", OutputTextFormat: api.TextAccented}}, false},
		{"Normalized", args{&api.TTSRequestConfig{Text: "aaa", OutputTextFormat: api.TextNormalized}}, false},
		{"Long", args{&api.TTSRequestConfig{Text: "111111111111111", OutputTextFormat: api.TextNone}}, false},
		{"tags", args{&api.TTSRequestConfig{Text: "aaa", OutputTextFormat: api.TextNone, SpeechMarkTypes: map[string]bool{"word": true}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.isOK(tt.args.inp); got != tt.want {
				t.Errorf("BigCacher.isOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_key(t *testing.T) {
	type args struct {
		inp *api.TTSRequestConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mp3", args{&api.TTSRequestConfig{Text: "olia", OutputFormat: api.AudioMP3}}, "olia_mp3_0.0000__0"},
		{"voice", args{&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioM4A,
			OutputTextFormat: api.TextAccented, Voice: "aaa"}}, "olia1_m4a_0.0000_aaa_0"},
		{"speed", args{&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioM4A,
			OutputTextFormat: api.TextAccented, Speed: 0.56, Voice: "aa"}}, "olia1_m4a_0.5600_aa_0"},
		{"test 2", args{&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioM4A,
			OutputTextFormat: api.TextAccented, Speed: 0.56, Voice: "aaa"}}, "olia1_m4a_0.5600_aaa_0"},
		{"max sil duration", args{&api.TTSRequestConfig{Text: "olia1", OutputFormat: api.AudioM4A,
			OutputTextFormat: api.TextAccented, Speed: 0.56, Voice: "aaa", MaxEdgeSilenceMillis: 50}}, "olia1_m4a_0.5600_aaa_50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := key(tt.args.inp); got != tt.want {
				t.Errorf("key() = %v, want %v", got, tt.want)
			}
		})
	}
}
