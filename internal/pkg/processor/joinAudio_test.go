package processor

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/wav"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewJoinAudio(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio()
	assert.NotNil(t, pr)
}

func TestJoinAudio(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio()
	d := synthesizer.TTSData{}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{&synthesizer.TTSDataPart{Audio: strA}}
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, strA, d.Audio)
}

func TestJoinAudio_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio()
	d := synthesizer.TTSData{}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{&synthesizer.TTSDataPart{Audio: strA}, &synthesizer.TTSDataPart{Audio: strA},
		&synthesizer.TTSDataPart{Audio: strA}}
	err := pr.Process(&d)
	assert.Nil(t, err)

	as := getTestAudioSize(strA)

	assert.Equal(t, as*3, getTestAudioSize(d.Audio))
}

func TestJoinAudio_DecodeFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio()
	d := synthesizer.TTSData{}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{&synthesizer.TTSDataPart{Audio: strA}, &synthesizer.TTSDataPart{Audio: "aaa"}}
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestJoinAudio_EmptyFail(t *testing.T) {
	initTestJSON(t)
	pr := NewJoinAudio()
	d := synthesizer.TTSData{}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{&synthesizer.TTSDataPart{Audio: strA}, &synthesizer.TTSDataPart{Audio: ""}}
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func getTestEncAudio(t *testing.T) string {
	return base64.StdEncoding.EncodeToString(getWaveData(t))
}

func getTestAudioSize(as string) uint32 {
	bt, _ := base64.StdEncoding.DecodeString(as)
	return wav.GetSize(bt)
}

func getWaveData(t *testing.T) []byte {
	res, err := ioutil.ReadFile("../wav/_testdata/test.wav")
	assert.Nil(t, err)
	return res
}
