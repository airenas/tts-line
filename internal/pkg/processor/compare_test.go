package processor

import (
	"context"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewComparator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewComparator("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewComparator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewComparator("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestCompareProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = "mod"
	d.PreviousText = "orig"
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*compOut) = compOut{BadAccents: []string{}, RC: 1}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp1 := httpJSONMock.Calls[0].Arguments[0]
	assert.Equal(t, &compIn{Original: "orig", Modified: "mod"}, cp1)
}

func TestCpmpareProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestCpmpareProcess_NoMatch(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*compOut) = compOut{BadAccents: []string{"a{x}"}, RC: 1}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Equal(t, utils.NewErrBadAccent([]string{"a{x}"}), err)
}

func TestCpmpareProcess_BadAccents(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*compOut) = compOut{BadAccents: []string{}, RC: 0}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Equal(t, utils.ErrTextDoesNotMatch, err)
}
