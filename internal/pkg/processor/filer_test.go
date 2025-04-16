package processor

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestNewFiler(t *testing.T) {
	pr, err := NewFiler("/dir")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestFilerSaves(t *testing.T) {
	pr, _ := NewFiler("/dir")
	assert.NotNil(t, pr)
	prf := pr.(*filer)
	b := &bytes.Buffer{}
	prf.fFile = func(name string) (io.WriteCloser, error) {
		assert.Equal(t, "/dir/out.mp3", name)
		return &testCloser{b}, nil
	}
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.AudioMP3 = []byte("mp3")
	err := prf.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", b.String())
}

func TestFiler_Skip(t *testing.T) {
	pr, _ := NewFiler("/dir")
	assert.NotNil(t, pr)
	prf := pr.(*filer)
	prf.fFile = func(name string) (io.WriteCloser, error) {
		assert.Fail(t, "Not expected file save")
		return nil, nil
	}
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioNone}}
	d.AudioMP3 = []byte("mp3")
	err := prf.Process(context.TODO(), &d)
	assert.Nil(t, err)
}

func TestFilerSaves_Fails(t *testing.T) {
	pr, _ := NewFiler("/dir")
	assert.NotNil(t, pr)
	prf := pr.(*filer)
	prf.fFile = func(name string) (io.WriteCloser, error) {
		assert.Equal(t, "/dir/out.mp3", name)
		return nil, errors.New("olia")
	}
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.AudioMP3 = []byte("mp3")
	err := prf.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestFilerSaves_FailsDecode(t *testing.T) {
	pr, _ := NewFiler("/dir")
	assert.NotNil(t, pr)
	d := synthesizer.TTSData{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}
	d.AudioMP3 = []byte("mp3")
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

type testCloser struct {
	io.Writer
}

func (mwc *testCloser) Close() error {
	return nil
}
