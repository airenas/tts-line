package acronyms

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/model"
	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	w1Mock *mocks.Worker
	w2Mock *mocks.Worker
)

func initTest(t *testing.T) {
	w1Mock = &mocks.Worker{}
	w2Mock = &mocks.Worker{}
}

func TestProcessFirst(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia"}}, nil)
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process(&model.Input{Word: "olia", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia", r[0].Word)
}

func TestProcessSecond(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, nil)
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process(&model.Input{Word: "olia", MI: "1"})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia2", r[0].Word)
}

func TestProcessLongWord(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, nil)
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process(&model.Input{Word: "olia1", MI: "X-"})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia1", r[0].Word)
}

func TestProcessShortWord(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, nil)
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process(&model.Input{Word: "olia", MI: "X-"})
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia2", r[0].Word)
}

func TestFailsFirst(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, errors.New("err"))
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	_, err = pr.Process(&model.Input{Word: "olia", MI: "1"})
	assert.NotNil(t, err)
}

func TestForceLetters(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, errors.New("err"))
	w2Mock.On("Process", mock.Anything, mock.Anything).Return([]api.ResultWord{{Word: "olia2"}}, nil) // should be called as letters
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	_, err = pr.Process(&model.Input{Word: "olia", MI: "1", ForceToLetters: true})
	assert.Nil(t, err)
}

func TestFailsSecond(t *testing.T) {
	initTest(t)
	w1Mock.On("Process", mock.Anything, mock.Anything).Return(nil, nil)
	w2Mock.On("Process", mock.Anything, mock.Anything).Return(nil, errors.New("err"))
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	_, err = pr.Process(&model.Input{Word: "olia", MI: "1"})
	assert.NotNil(t, err)
}

func TestAsLetters(t *testing.T) {
	assert.True(t, canReadAsLetters("aaa"))
	assert.True(t, canReadAsLetters("aaaa"))
	assert.False(t, canReadAsLetters("aaaaa"))
	assert.True(t, canReadAsLetters("lrt.lt"))
	assert.False(t, canReadAsLetters("lrtas.lt"))
	assert.True(t, canReadAsLetters("lrt.eu"))
	assert.False(t, canReadAsLetters("lrt.va"))
}
