package processor

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/audio"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/internal/pkg/utils"
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
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, getWaveData(t), d.Audio)
	assert.InDelta(t, 0.5572, d.AudioLenSeconds, 0.001)
}

func TestJoinAudio_Skip(t *testing.T) {
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Nil(t, d.Audio)
	assert.InDelta(t, 0.0, d.AudioLenSeconds, 0.001)
}

func TestJoinAudio_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: strA},
		{Audio: strA}}
	err := pr.Process(context.TODO(), &d)
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
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: []byte("aaa")}}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestJoinAudio_EmptyFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: []byte("")}}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestJoinAudio_SuffixFail(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}, AudioSuffix: "test.wav"}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	loaderMock.On("TakeWav", mock.Anything).Return(nil, errors.New("fail"))
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestJoinAudio_Suffix(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}, AudioSuffix: "test.wav"}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	loaderMock.On("TakeWav", mock.Anything).Return(getWaveData(t), nil)
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, 0.5572*2, d.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(context.TODO(), da)
	assert.Nil(t, err)
	assert.Equal(t, getWaveData(t), da.Audio)
	assert.InDelta(t, 0.5572, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_Skip(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Cfg: &synthesizer.TTSConfig{}}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(context.TODO(), da)
	assert.Nil(t, err)
	assert.Nil(t, da.Audio)
	assert.InDelta(t, 0.0, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}, {Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}},
		{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d, &d}}
	err := pr.Process(context.TODO(), da)
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
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Cfg: &synthesizer.TTSConfig{}}, {Audio: []byte("aaa"), Cfg: &synthesizer.TTSConfig{}}}
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(context.TODO(), da)
	assert.NotNil(t, err)
}

func TestJoinSSMLAudio_EmptyFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}, {Audio: []byte(""), Step: 256, Cfg: &synthesizer.TTSConfig{}}}
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}}
	err := pr.Process(context.TODO(), da)
	assert.NotNil(t, err)
}

func TestJoinSSMLAudio_AddPause(t *testing.T) {
	pr := NewJoinSSMLAudio(loaderMock)
	d := &synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}}
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
		{name: "pause start", args: []*synthesizer.TTSData{dp, d}, wantSize: as + ps, wantLen: al + 5, wantErr: false},
		{name: "simple", args: []*synthesizer.TTSData{d}, wantSize: as, wantLen: al, wantErr: false},
		{name: "pause end", args: []*synthesizer.TTSData{d, dp}, wantSize: as + ps, wantLen: al + 5, wantErr: false},
		{name: "pause middle", args: []*synthesizer.TTSData{d, dp, d}, wantSize: as*2 + ps, wantLen: al*2 + 5, wantErr: false},
		{name: "max 10 sec", args: []*synthesizer.TTSData{d, dp, dp, dp}, wantSize: as + ps*2, wantLen: al + 10, wantErr: false},
		{name: "just pause", args: []*synthesizer.TTSData{dp, dp, dp}, wantSize: 0, wantLen: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &synthesizer.TTSData{Input: d.Input, SSMLParts: tt.args}
			err := pr.Process(context.TODO(), da)
			require.Equal(t, tt.wantErr, err != nil)
			if !tt.wantErr {
				assert.InDelta(t, int(tt.wantSize), int(getTestAudioSize(da.Audio)), 300) // give some delta bacause of hops to time conversion and floating point errors
				assert.InDelta(t, tt.wantLen, da.AudioLenSeconds, 0.004)
			}
		})
	}

}

func TestJoinSSMLAudio_Suffix(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}, AudioSuffix: "oo.wav"}
	loaderMock.On("TakeWav", mock.Anything).Return(getWaveData(t), nil)
	err := pr.Process(context.TODO(), da)
	assert.Nil(t, err)
	assert.InDelta(t, 0.5572*2, da.AudioLenSeconds, 0.001)
}

func TestJoinSSMLAudio_SuffixFail(t *testing.T) {
	initTestJoiner(t)
	pr := NewJoinSSMLAudio(loaderMock)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA, Step: 256, Cfg: &synthesizer.TTSConfig{}}}
	d.Cfg.Type = synthesizer.SSMLText
	da := &synthesizer.TTSData{Input: d.Input, SSMLParts: []*synthesizer.TTSData{&d}, AudioSuffix: "oo.wav"}
	loaderMock.On("TakeWav", mock.Anything).Return(nil, errors.New("fail"))
	err := pr.Process(context.TODO(), da)
	assert.NotNil(t, err)
}

func getTestEncAudio(t *testing.T) []byte {
	t.Helper()
	return getWaveData(t)
}

func getTestAudioSize(bt []byte) uint32 {
	return wav.GetSize(bt)
}

func getTestPauseSize(bt []byte, dur time.Duration) uint32 {
	sr := wav.GetSampleRate(bt)
	bps := wav.GetBitsPerSample(bt)
	return uint32(dur.Milliseconds() * int64(sr) * int64(bps) / (8 * 1000))
}

func getWaveData(t *testing.T) []byte {
	t.Helper()
	return getWaveDataWithName(t, "test.wav")
}

func getWaveDataWithName(tb testing.TB, filename string) []byte {
	tb.Helper()
	res, err := os.ReadFile("../wav/_testdata/" + filename)
	require.Nil(tb, err)
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
			got, _ := writePause(context.TODO(), &buf, tt.args.sampleRate, tt.args.bitsPerSample, tt.args.pause)
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

func Test_calcPauseWithEnds(t *testing.T) {
	type args struct {
		s1    int
		s2    int
		pause int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 int
	}{
		{name: "no data", args: args{s1: 0, s2: 0, pause: 0}, want: 0, want1: 0, want2: 0},
		{name: "returns pause", args: args{s1: 0, s2: 0, pause: 10}, want: 0, want1: 0, want2: 10},
		{name: "returns s1", args: args{s1: 5, s2: 0, pause: 10}, want: 0, want1: 0, want2: 5},
		{name: "returns s2", args: args{s1: 0, s2: 5, pause: 10}, want: 0, want1: 0, want2: 5},
		{name: "returns s2", args: args{s1: 5, s2: 5, pause: 10}, want: 0, want1: 0, want2: 0},
		{name: "returns s2", args: args{s1: 6, s2: 6, pause: 10}, want: 1, want1: 1, want2: 0},
		{name: "returns s2", args: args{s1: 7, s2: 6, pause: 10}, want: 2, want1: 1, want2: 0},
		{name: "returns s2", args: args{s1: 6, s2: 7, pause: 10}, want: 1, want1: 2, want2: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := calcPauseWithEnds(tt.args.s1, tt.args.s2, tt.args.pause)
			if got != tt.want {
				t.Errorf("calcPauseWithEnds() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("calcPauseWithEnds() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("calcPauseWithEnds() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_fixPause(t *testing.T) {
	type args struct {
		s1    int
		s2    int
		pause time.Duration
		step  int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 time.Duration
	}{
		{name: "leaves pause", args: args{s1: 10, s2: 3, pause: time.Second * 1, step: 2205}, want: 3, want1: 0, want2: time.Second * 0},
		{name: "no data", args: args{s1: 1, s2: 2, pause: 10, step: 0}, want: 1, want1: 2, want2: 10},
		{name: "no drop pause", args: args{s1: 10, s2: 10, pause: time.Second, step: 2205}, want: 5, want1: 5, want2: 0},
		{name: "leaves pause", args: args{s1: 10, s2: 10, pause: time.Second * 5, step: 2205}, want: 0, want1: 0, want2: time.Second * 3},
		// 1s - (30*256 + 20*256)/22050
		{name: "leaves pause", args: args{s1: 30, s2: 20, pause: time.Second * 1, step: 256}, want: 0, want1: 0, want2: time.Millisecond * 417},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := fixPause(tt.args.s1, tt.args.s2, tt.args.pause, tt.args.step, 22050)
			if got != tt.want {
				t.Errorf("fixPause() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("fixPause() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("fixPause() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_getStartSilSize(t *testing.T) {
	type args struct {
		phones    []string
		durations []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "none", args: args{phones: []string{}, durations: []int{}}, want: 0},
		{name: "none", args: args{phones: []string{"sil", "a", "a", "a", "a", "a", "a"}, durations: []int{}}, want: 0},
		{name: "none", args: args{phones: []string{"a", "b"}, durations: []int{1, 2, 3}}, want: 0},
		{name: "finds", args: args{phones: []string{"sil", "a", "b", "sp", "sil"}, durations: []int{10, 2, 3, 5, 6, 7}}, want: 10},
		{name: "finds", args: args{phones: []string{"sil", "-", "a", "b", "sp", "sil"}, durations: []int{10, 2, 3, 5, 6, 7}}, want: 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStartSilSize(tt.args.phones, tt.args.durations); got != tt.want {
				t.Errorf("getStartSilSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEndSilSize(t *testing.T) {
	type args struct {
		phones    []string
		durations []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "none", args: args{phones: []string{}, durations: []int{}}, want: 0},
		{name: "none", args: args{phones: []string{"sil", "a", "a", "a", "a", "a", "a", "sil"}, durations: []int{}}, want: 0},
		{name: "none", args: args{phones: []string{"a", "b"}, durations: []int{1, 2, 3}}, want: 3},
		{name: "finds", args: args{phones: []string{"sil", "a", "b", "sp", "sil"}, durations: []int{10, 2, 3, 5, 6, 7}}, want: 18},
		{name: "finds", args: args{phones: []string{"sil", "-", "a", "b", ",", "sp", "sil"}, durations: []int{10, 2, 3, 5, 6, 7, 8, 9}}, want: 30},
		{name: "finds punct", args: args{phones: []string{"sil", "-", "a", "-", ".", ",", "sp", "sil"}, durations: []int{10, 2, 3, 5, 6, 7, 8, 9, 10}}, want: 45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEndSilSize(context.TODO(), tt.args.phones, tt.args.durations); got != tt.want {
				t.Errorf("getEndSilSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSil(t *testing.T) {
	type args struct {
		ph string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "punc", args: args{ph: "."}, want: true},
		{name: "not", args: args{ph: "a"}, want: false},
		{name: "not", args: args{ph: "^a"}, want: false},
		{name: "not", args: args{ph: "\"a"}, want: false},
		{name: "punc", args: args{ph: "."}, want: true},
		{name: "punc", args: args{ph: ","}, want: true},
		{name: "punc", args: args{ph: ":"}, want: true},
		{name: "punc", args: args{ph: "?"}, want: true},
		{name: "punc", args: args{ph: "!"}, want: true},
		{name: "punc", args: args{ph: ";"}, want: true},
		{name: "sil", args: args{ph: "sil"}, want: true},
		{name: "sp", args: args{ph: "sp"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSil(tt.args.ph); got != tt.want {
				t.Errorf("isSil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toHops(t *testing.T) {
	type args struct {
		millis     int64
		step       int
		sampleRate uint32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "1s", args: args{millis: 1000, step: 256, sampleRate: 22050}, want: 86},
		{name: "0.5s", args: args{millis: 500, step: 256, sampleRate: 22050}, want: 43},
		{name: "0s", args: args{millis: 0, step: 256, sampleRate: 22050}, want: 0},
		{name: "10ms", args: args{millis: 10, step: 256, sampleRate: 22050}, want: 1},
		{name: "40ms round", args: args{millis: 40, step: 256, sampleRate: 22050}, want: 3},
		{name: "30ms round", args: args{millis: 30, step: 256, sampleRate: 22050}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toHops(tt.args.millis, tt.args.step, tt.args.sampleRate); got != tt.want {
				t.Errorf("toHops() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calcVolumeRate(t *testing.T) {
	tests := []struct {
		name       string
		changeInDB float64
		want       float64
	}{
		{name: "no change", changeInDB: 0.0, want: 1.0},
		{name: "increase 6dB", changeInDB: 6.0, want: 1.9952623149688795},
		{name: "decrease 6dB", changeInDB: -6.0, want: 0.5011872336272722},
		{name: "increase 20dB", changeInDB: 20.0, want: 10.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcVolumeRate(tt.changeInDB)
			if !utils.Float64Equals(got, tt.want) {
				t.Errorf("calcVolumeRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixStartEndRates(t *testing.T) {
	tests := []struct {
		name        string
		res         []*audio.VolChange
		defaultGain float64
		want        []*audio.VolChange
	}{
		{name: "nil", res: nil, defaultGain: 1.0, want: nil},
		{name: "empty", res: []*audio.VolChange{}, defaultGain: 1.0, want: []*audio.VolChange{}},
		{name: "one", res: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5}}, defaultGain: 1.2,
			want: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5, StartRate: 1.2, EndRate: 1.2}}},
		{name: "two in a row", res: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5}, {From: 2000, To: 3000, Rate: 2}}, defaultGain: 1.2,
			want: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5, StartRate: 1.2, EndRate: 1.5},
				{From: 2000, To: 3000, Rate: 2, StartRate: 1.5, EndRate: 1.2}}},
		{name: "not in a row", res: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5}, {From: 3000, To: 4000, Rate: 2}}, defaultGain: 1.2,
			want: []*audio.VolChange{{From: 1000, To: 2000, Rate: 1.5, StartRate: 1.2, EndRate: 1.2},
				{From: 3000, To: 4000, Rate: 2, StartRate: 1.2, EndRate: 1.2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixStartEndRates(tt.res, tt.defaultGain)
			if !reflect.DeepEqual(toNonPtrArray(got), toNonPtrArray(tt.want)) {
				t.Errorf("fixStartEndRates() = %v, want %v", toNonPtrArray(got), toNonPtrArray(tt.want))
			}
		})
	}
}

func toNonPtrArray[T any](in []*T) []T {
	if in == nil {
		return nil
	}
	res := make([]T, 0, len(in))
	for _, v := range in {
		res = append(res, *v)
	}
	return res
}
