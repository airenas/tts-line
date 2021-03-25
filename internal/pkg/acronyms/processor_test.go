package acronyms

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	w1Mock *mocks.MockWorker
	w2Mock *mocks.MockWorker
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	w1Mock = mocks.NewMockWorker()
	w2Mock = mocks.NewMockWorker()
}

func TestProcessFirst(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia"}}, nil)
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process("olia", "1")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia", r[0].Word)
}

func TestProcessSecond(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, nil)
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process("olia", "1")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia2", r[0].Word)
}

func TestProcessLongWord(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, nil)
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process("olia1", "X-")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia1", r[0].Word)
}

func TestProcessShortWord(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, nil)
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	r, err := pr.Process("olia", "X-")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia2", r[0].Word)
}

func TestFailsFirst(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn([]api.ResultWord{{Word: "olia2"}}, nil)
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	_, err = pr.Process("olia", "1")
	assert.NotNil(t, err)
}

func TestFailsSecond(t *testing.T) {
	initTest(t)
	pegomock.When(w1Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, nil)
	pegomock.When(w2Mock.Process(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	pr, err := NewProcessor(w1Mock, w2Mock)
	assert.Nil(t, err)
	assert.NotNil(t, pr)
	_, err = pr.Process("olia", "1")
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
