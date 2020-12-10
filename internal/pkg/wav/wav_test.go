package wav

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTakeHeader(t *testing.T) {
	res := TakeHeader(testWaveData(t))
	assert.Equal(t, "RIFF$\xc0\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00D\xac\x00\x00\x10\xb1\x02\x00\x02\x00\x10\x00data", string(res))
}

func TestTakeData(t *testing.T) {
	res := TakeData(testWaveData(t))
	assert.Equal(t, "4da9eb077b0af2afee01ef5c15408371", fmt.Sprintf("%x", md5.Sum(res)))
}

func TestGetSize(t *testing.T) {
	res := GetSize(testWaveData(t))
	assert.Equal(t, uint32(49152), res)
}

func TestSizeBytes(t *testing.T) {
	assert.Equal(t, "00000000", fmt.Sprintf("%x", SizeBytes(0)))
	assert.Equal(t, "01000000", fmt.Sprintf("%x", SizeBytes(1)))
	assert.Equal(t, "10000000", fmt.Sprintf("%x", SizeBytes(16)))
	assert.Equal(t, "00010000", fmt.Sprintf("%x", SizeBytes(256)))
}

func testWaveData(t *testing.T) []byte {
	res, err := ioutil.ReadFile("_testdata/test.wav")
	assert.Nil(t, err)
	return res
}
