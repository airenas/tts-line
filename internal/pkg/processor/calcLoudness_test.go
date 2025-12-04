package processor

import (
	"context"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
)

func TestCalcLoudness(t *testing.T) {
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getWaveDataWithName(t, "sine_1s.wav")
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: strA}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, -21.70, d.Parts[0].Loudness, 1) // very quiet with zeros
	assert.InDelta(t, 0.0, d.Parts[0].LoudnessGain, 0.001)
}

func TestCalcLoudness_Skip(t *testing.T) {
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	strA := getWaveDataWithName(t, "sine_1s.wav")
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: strA}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, 0.0, d.Parts[0].Loudness, 0.001)
	assert.InDelta(t, 0.0, d.Parts[0].LoudnessGain, 0.001)
}
func TestCalcLoudness_SkipOne(t *testing.T) {
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getWaveDataWithName(t, "sine_1s.wav")
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, 0.0, d.Parts[0].Loudness, 1)
	assert.InDelta(t, 0.0, d.Parts[0].LoudnessGain, 0.001)
}

func TestCalcLoudness_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.Parts = []*synthesizer.TTSDataPart{{Audio: getWaveDataWithName(t, "sine_1s.wav")}, {Audio: getWaveDataWithName(t, "sine_louder_1s.wav")}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, -21.70, d.Parts[0].Loudness, 1)
	assert.InDelta(t, -15.70, d.Parts[1].Loudness, 1)
	assert.InDelta(t, 0, d.Parts[0].LoudnessGain, 0.1)
	assert.InDelta(t, -6, d.Parts[1].LoudnessGain, 0.1)
}

func TestCalcLoudness_Several_Increase(t *testing.T) {
	initTestJSON(t)
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.Parts = []*synthesizer.TTSDataPart{{Audio: getWaveDataWithName(t, "sine_louder_1s.wav")}, {Audio: getWaveDataWithName(t, "sine_1s.wav")}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, -15.70, d.Parts[0].Loudness, 1)
	assert.InDelta(t, -21.70, d.Parts[1].Loudness, 1)
	assert.InDelta(t, 0, d.Parts[0].LoudnessGain, 0.1)
	assert.InDelta(t, 6, d.Parts[1].LoudnessGain, 0.1)
}

func TestCalcLoudness_SkipZeros(t *testing.T) {
	initTestJSON(t)
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.Parts = []*synthesizer.TTSDataPart{
		{Audio: getWaveDataWithName(t, "test.wav")},
		{Audio: getWaveDataWithName(t, "sine_louder_1s.wav")},
		{Audio: getWaveDataWithName(t, "sine_1s.wav")}}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.InDelta(t, -70, d.Parts[0].Loudness, 1)
	assert.InDelta(t, -15.70, d.Parts[1].Loudness, 1)
	assert.InDelta(t, -21.70, d.Parts[2].Loudness, 1)
	assert.InDelta(t, 0, d.Parts[0].LoudnessGain, 0.1)
	assert.InDelta(t, 0, d.Parts[1].LoudnessGain, 0.1)
	assert.InDelta(t, 6, d.Parts[2].LoudnessGain, 0.1)
}

func TestCalcLoudness_WavFail(t *testing.T) {
	pr := NewCalcLoudness(2)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	strA := getTestEncAudio(t)
	d.Parts = []*synthesizer.TTSDataPart{{Audio: strA}, {Audio: []byte("aaa")}}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestCalcLoudnessSSML_Several(t *testing.T) {
	initTestJSON(t)
	pr := NewCalcLoudnessSSML(2)
	d1 := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d1.Parts = []*synthesizer.TTSDataPart{{Audio: getWaveDataWithName(t, "sine_louder_1s.wav")}, {Audio: getWaveDataWithName(t, "sine_1s.wav")}}
	d2 := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d2.Parts = []*synthesizer.TTSDataPart{{Audio: getWaveDataWithName(t, "sine_louder_1s.wav")}, {Audio: getWaveDataWithName(t, "sine_1s.wav")}}

	input := synthesizer.TTSData{
		SSMLParts: []*synthesizer.TTSData{&d1, &d2},
		Input:     &api.TTSRequestConfig{OutputFormat: api.AudioMP3},
	}
	err := pr.Process(context.TODO(), &input)
	assert.Nil(t, err)
	assert.InDelta(t, -15.70, d1.Parts[0].Loudness, 1)
	assert.InDelta(t, -21.70, d1.Parts[1].Loudness, 1)
	assert.InDelta(t, 0, d1.Parts[0].LoudnessGain, 0.1)
	assert.InDelta(t, 6, d1.Parts[1].LoudnessGain, 0.1)
	assert.InDelta(t, -15.70, d2.Parts[0].Loudness, 1)
	assert.InDelta(t, -21.70, d2.Parts[1].Loudness, 1)
	assert.InDelta(t, 0, d2.Parts[0].LoudnessGain, 0.1)
	assert.InDelta(t, 6, d2.Parts[1].LoudnessGain, 0.1)
}

func TestCalcLoudnessSSML_Skip(t *testing.T) {
	initTestJSON(t)
	pr := NewCalcLoudnessSSML(2)
	d1 := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d1.Parts = []*synthesizer.TTSDataPart{}
	d2 := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d2.Parts = []*synthesizer.TTSDataPart{{Audio: getWaveDataWithName(t, "sine_louder_1s.wav")}}

	input := synthesizer.TTSData{
		SSMLParts: []*synthesizer.TTSData{&d1, &d2},
		Input:     &api.TTSRequestConfig{OutputFormat: api.AudioMP3},
	}
	err := pr.Process(context.TODO(), &input)
	assert.Nil(t, err)
	assert.InDelta(t, 0, d2.Parts[0].Loudness, 1)
	assert.InDelta(t, 0, d2.Parts[0].LoudnessGain, 0.1)
}
