package wav

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTakeHeader(t *testing.T) {
	res := TakeHeader(getWaveData(t))
	assert.Equal(t, "RIFF$\xc0\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00D\xac\x00\x00\x10\xb1\x02\x00\x02\x00\x10\x00data", string(res))
}

func TestIsValid(t *testing.T) {
	res := getWaveData(t)
	assert.True(t, IsValid(res))
	assert.True(t, IsValid(res[:50]))
	assert.False(t, IsValid(res[:40]))
	assert.False(t, IsValid(nil))
}

func TestTakeData(t *testing.T) {
	res := TakeData(getWaveData(t))
	assert.Equal(t, "4da9eb077b0af2afee01ef5c15408371", fmt.Sprintf("%x", md5.Sum(res)))
}

func TestGetSize(t *testing.T) {
	res := GetSize(getWaveData(t))
	assert.Equal(t, uint32(49152), res)
}

func TestSizeBytes(t *testing.T) {
	assert.Equal(t, "00000000", fmt.Sprintf("%x", SizeBytes(0)))
	assert.Equal(t, "01000000", fmt.Sprintf("%x", SizeBytes(1)))
	assert.Equal(t, "10000000", fmt.Sprintf("%x", SizeBytes(16)))
	assert.Equal(t, "00010000", fmt.Sprintf("%x", SizeBytes(256)))
}

func TestSampleRate(t *testing.T) {
	wave := getWaveData(t)
	assert.Equal(t, 44100, int(GetSampleRate(wave)))
	wave = getWaveDataN(t, "test2")
	assert.Equal(t, 44100, int(GetSampleRate(wave)))
}

func TestChannels(t *testing.T) {
	wave := getWaveData(t)
	assert.Equal(t, 1, int(GetChannels(wave)))
	wave = getWaveDataN(t, "test2")
	assert.Equal(t, 2, int(GetChannels(wave)))
}

func TestGetBitsPerSample(t *testing.T) {
	wave := getWaveData(t)
	assert.Equal(t, 16, int(GetBitsPerSample(wave)))
	wave = getWaveDataN(t, "test2")
	assert.Equal(t, 16, int(GetBitsPerSample(wave)))
}

func TestLen(t *testing.T) {
	wave := getWaveData(t)
	assert.InDelta(t, 0.2786, float64(GetSize(wave))/float64(GetBitsRate(wave)), 0.001)
	assert.InDelta(t, .557, float64(GetSize(wave))/float64(GetBitsRateCalc(wave)), 0.001)
	wave = getWaveDataN(t, "test2")
	assert.InDelta(t, .557, float64(GetSize(wave))/float64(GetBitsRate(wave)), 0.001)
	assert.InDelta(t, .557, float64(GetSize(wave))/float64(GetBitsRateCalc(wave)), 0.001)
}

func getWaveData(t *testing.T) []byte {
	return getWaveDataN(t, "test")
}

func getWaveDataN(t *testing.T, name string) []byte {
	res, err := ioutil.ReadFile(fmt.Sprintf("_testdata/%s.wav", name))
	assert.Nil(t, err)
	return res
}
