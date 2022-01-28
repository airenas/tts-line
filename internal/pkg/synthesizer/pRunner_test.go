package synthesizer

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

var (
	partProcTest *partProcMock
	runner       *PartRunner
	d            *TTSData
)

func initPRunnerTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	partProcTest = &partProcMock{f: func(d *TTSDataPart) error { return nil }}
	runner = NewPartRunner(0)
	runner.Add(partProcTest)
	d = &TTSData{}
}

func TestPRunner_Default(t *testing.T) {
	runner = NewPartRunner(0)
	assert.Equal(t, 3, runner.parallelWorker)
	runner = NewPartRunner(10)
	assert.Equal(t, 10, runner.parallelWorker)
}
func TestPRProcess(t *testing.T) {
	initPRunnerTest(t)
	partProcTest.f = func(d *TTSDataPart) error {
		assert.Equal(t, "olia", d.Spectogram)
		d.Audio = "mp3"
		return nil
	}
	d.Parts = append(d.Parts, &TTSDataPart{Spectogram: "olia"})
	err := runner.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", d.Parts[0].Audio)
}

func TestPRProcess_Several(t *testing.T) {
	initPRunnerTest(t)
	partProcTest.f = func(d *TTSDataPart) error {
		d.Audio = "mp3"
		return nil
	}
	d.Parts = append(d.Parts, &TTSDataPart{})
	d.Parts = append(d.Parts, &TTSDataPart{})
	err := runner.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "mp3", d.Parts[0].Audio)
	assert.Equal(t, "mp3", d.Parts[1].Audio)
}

func TestPRProcess_Fail(t *testing.T) {
	initPRunnerTest(t)
	partProcTest.f = func(d *TTSDataPart) error {
		return errors.New("error")
	}
	d.Parts = append(d.Parts, &TTSDataPart{})
	d.Parts = append(d.Parts, &TTSDataPart{})
	err := runner.Process(d)
	assert.NotNil(t, err)
}

func TestPRProcess_StopOnError(t *testing.T) {
	initPRunnerTest(t)
	runner.parallelWorker = 1
	c := int32(0)
	partProcTest.f = func(d *TTSDataPart) error {
		atomic.AddInt32(&c, 1)
		return errors.New("error")
	}
	d.Parts = append(d.Parts, &TTSDataPart{})
	d.Parts = append(d.Parts, &TTSDataPart{})
	err := runner.Process(d)
	assert.NotNil(t, err)
	assert.Equal(t, int32(1), c)
}

func TestPRProcess_StopSeveralProcessors(t *testing.T) {
	initPRunnerTest(t)
	runner.parallelWorker = 1
	c := int32(0)
	partProcTest.f = func(d *TTSDataPart) error {
		atomic.AddInt32(&c, 1)
		return errors.New("error")
	}

	partProcTest1 := &partProcMock{f: func(d *TTSDataPart) error {
		atomic.AddInt32(&c, 1)
		return nil
	}}
	runner.Add(partProcTest1)

	d.Parts = append(d.Parts, &TTSDataPart{})
	d.Parts = append(d.Parts, &TTSDataPart{})
	err := runner.Process(d)
	assert.NotNil(t, err)
	assert.Equal(t, int32(1), c)
}

func TestPRProcess_StopSeveralProcessors2(t *testing.T) {
	initPRunnerTest(t)
	runner.parallelWorker = 2
	c := int32(0)
	partProcTest.f = func(d *TTSDataPart) error {
		atomic.AddInt32(&c, 1)
		time.Sleep(10 * time.Millisecond)
		return errors.New("error")
	}

	partProcTest1 := &partProcMock{f: func(d *TTSDataPart) error {
		atomic.AddInt32(&c, 1)
		return nil
	}}
	runner.Add(partProcTest1)

	d.Parts = append(d.Parts, &TTSDataPart{})
	d.Parts = append(d.Parts, &TTSDataPart{})
	err := runner.Process(d)
	assert.NotNil(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&c))
}

type partProcMock struct {
	f func(res *TTSDataPart) error
}

func (pr *partProcMock) Process(d *TTSDataPart) error {
	return pr.f(d)
}
