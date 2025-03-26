package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
)

func TestSSMLPartRunner_InvokesTextParts(t *testing.T) {
	c := 0
	m := procMock{f: func(ctx context.Context, res *synthesizer.TTSData) error { c++; return nil }}
	r := NewSSMLPartRunner([]synthesizer.Processor{&m})
	d := synthesizer.TTSData{SSMLParts: []*synthesizer.TTSData{{}, {}, {}}}
	d.SSMLParts[0].Cfg.Type = synthesizer.SSMLText
	d.SSMLParts[1].Cfg.Type = synthesizer.SSMLText
	err := r.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, 2, c)
	r = NewSSMLPartRunner([]synthesizer.Processor{&m, &m})
	c = 0
	err = r.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, 4, c)
}

func TestSSMLPartRunner_Fail(t *testing.T) {
	m := procMock{f: func(ctx context.Context, res *synthesizer.TTSData) error { return errors.New("fail") }}
	r := NewSSMLPartRunner([]synthesizer.Processor{&m})
	d := synthesizer.TTSData{SSMLParts: []*synthesizer.TTSData{{}, {}, {}}}
	d.SSMLParts[0].Cfg.Type = synthesizer.SSMLText
	err := r.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

type procMock struct {
	f func(ctx context.Context, res *synthesizer.TTSData) error
}

func (pr *procMock) Process(ctx context.Context, d *synthesizer.TTSData) error {
	return pr.f(ctx, d)
}
