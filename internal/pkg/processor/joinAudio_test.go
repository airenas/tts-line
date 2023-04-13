package processor

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/wav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	loaderMock *mockAudioLoader
)

func initTestJoiner(t *testing.T) {
	t.Helper()
	loaderMock = &mockAudioLoader{}
}

func TestNewJoinAudio(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinAudio(loaderMock)
	assert.NotNil(t, pr)
}

func TestJoinAudio(t *testing.T) {
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, strA, d.Audio)
	assert.InDelta(t, 0.5572, d.AudioLenSeconds, 0.001)
}

func TestJoinAudio_Skip(t *testing.T) {
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "", d.Audio)
	assert.InDelta(t, 0.0, d.AudioLenSeconds, 0.001)
}

func TestJoinAudio_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: strA},
		{Audio: strA}}
	err := pr.Process(&d)
	assert.Nil(t, err)

	as := getTestAudioSize(strA)

	assert.Equal(t, as*3, getTestAudioSize(d.Audio))
	assert.InDelta(t, 0.5572*3, d.AudioLenSeconds, 0.001)
}

func TestJoinAudio_DecodeFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: "aaa"}}
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestJoinAudio_EmptyFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: ""}}
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestJoinAudio_SuffixFail(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}, AudioSuffix: "test.wav"}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	loaderMock.On("TakeWav", mock.Anything).Return(nil, errors.New("fail"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestJoinAudio_Suffix(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}, AudioSuffix: "test.wav"}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	loaderMock.On("TakeWav", mock.Anything).Return(getWaveData(t), nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.InDelta(t, 0.5572*2, d.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(da)
	assert.Nil(t, err)
	assert.Equal(t, strA, da.Audio)
	assert.InDelta(t, 0.5572, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_Skip(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(da)
	assert.Nil(t, err)
	assert.Equal(t, "", da.Audio)
	assert.InDelta(t, 0.0, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: strA},
		{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d, &d}}
	err := pr.Process(da)
	assert.Nil(t, err)

	as := getTestAudioSize(strA)

	assert.Equal(t, as*6, getTestAudioSize(da.Audio))
	assert.InDelta(t, 0.5572*6, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_DecodeFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: "aaa"}}
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(da)
	assert.NotNil(t, err)
}

func TestJoinSSMLAudio_EmptyFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: ""}}
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(da)
	assert.NotNil(t, err)
}

func TestJoinSSMLAudio_AddPause(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := &synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	dp := &synthesizer.TTSData{}
	dp.Cfg.Type = synthesizer.SSMLPause
	dp.Cfg.PauseDuration = time.Second * 5

	as := getTestAudioSize(strA)
	ps := getTestPauseSize(strA, dp.Cfg.PauseDuration)
	al := 0.5572

	tests := []struct {
		name     string
		args     []*synthesizer.TTSData
		wantSize uint32
		wantLen  float64
		wantErr  bool
	}{
		{name: "simple", args: []*synthesizer.TTSData{d}, wantSize: as, wantLen: al, wantErr: false},
		{name: "pause start", args: []*synthesizer.TTSData{dp, d}, wantSize: as + ps, wantLen: al + 5, wantErr: false},
		{name: "pause end", args: []*synthesizer.TTSData{d, dp}, wantSize: as + ps, wantLen: al + 5, wantErr: false},
		{name: "pause middle", args: []*synthesizer.TTSData{d, dp, d}, wantSize: as*2 + ps, wantLen: al*2 + 5, wantErr: false},
		{name: "max 10 sec", args: []*synthesizer.TTSData{d, dp, dp, dp}, wantSize: as + ps*2, wantLen: al + 10, wantErr: false},
		{name: "just pause", args: []*synthesizer.TTSData{dp, dp, dp}, wantSize: 0, wantLen: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &synthesizer.TTSData{Input: d.Input, SSMLParts: tt.args}
			err := pr.Process(da)
			require.Equal(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				assert.Equal(t, tt.wantSize, getTestAudioSize(da.Audio))
				assert.InDelta(t, tt.wantLen, da.AudioLenSeconds, 0.001)
			}
		})
	}

}

func TestJoinSSMLAudio_Suffix(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}, AudioSuffix: "oo.wav"}
	loaderMock.On("TakeWav", mock.Anything).Return(getWaveData(t), nil)
	err := pr.Process(da)
	assert.Nil(t, err)
	assert.InDelta(t, 0.5572*2, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_SuffixFail(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}, AudioSuffix: "oo.wav"}
	loaderMock.On("TakeWav", mock.Anything).Return(nil, errors.New("fail"))
	err := pr.Process(da)
	assert.NotNil(t, err)
}

func getTestEncAudio(t *testing.T) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString(getWaveData(t))
}

func getTestAudioSize(as string) uint32 {
	bt, _ := base64.StdEncoding.DecodeString(as)
	return wav.GetSize(bt)
}

func getTestPauseSize(as string, dur time.Duration) uint32 {
	bt, _ := base64.StdEncoding.DecodeString(as)
	sr := wav.GetSampleRate(bt)
	bps := wav.GetBitsPerSample(bt)
	return uint32(dur.Milliseconds() * int64(sr) * int64(bps) / (8 * 1000))
}

func getWaveData(t *testing.T) []byte {
	t.Helper()
	res, err := os.ReadFile("../wav/_testdata/test.wav")
	assert.Nil(t, err)
	return res
}

func Test_writePause(t *testing.T) {
	type args struct {
		sampleRate    uint32
		bitsPerSample uint16
		pause         time.Duration
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "simple", args: args{sampleRate: 22050, bitsPerSample: 16, pause: time.Second}, want: 22050 * 2},
		{name: "8 bits", args: args{sampleRate: 22050, bitsPerSample: 8, pause: time.Second}, want: 22050},
		{name: "24 bits", args: args{sampleRate: 22050, bitsPerSample: 24, pause: time.Second}, want: 22050 * 3},
		{name: "9s", args: args{sampleRate: 22050, bitsPerSample: 16, pause: time.Second * 9}, want: 22050 * 2 * 9},
		{name: "15s", args: args{sampleRate: 22050, bitsPerSample: 16, pause: time.Second * 15}, want: 22050 * 2 * 10},
		{name: "-1s", args: args{sampleRate: 22050, bitsPerSample: 16, pause: -time.Second * 9}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			got, _ := writePause(&buf, tt.args.sampleRate, tt.args.bitsPerSample, tt.args.pause)
			assert.Equal(t, tt.want, got)
			require.Equal(t, int(tt.want), len(buf.Bytes()))
			for _, b := range buf.Bytes() {
				require.Equal(t, byte(0), b)
			}
		})
	}
}

type mockAudioLoader struct{ mock.Mock }

func (m *mockAudioLoader) TakeWav(in string) ([]byte, error) {
	args := m.Called(in)
	return mocks.To[[]byte](args.Get(0)), args.Error(1)
}
