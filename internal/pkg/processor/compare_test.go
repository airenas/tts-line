package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*compOut) = compOut{BadAccents: []string{}, RC: 1}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Nil(t, err)
	cp1, _ := httpJSONMock.VerifyWasCalledOnce().InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface()).GetCapturedArguments()
	assert.Equal(t, &compIn{Original: "orig", Modified: "mod"}, cp1)
}

func TestCpmpareProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).ThenReturn(errors.New("olia"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestCpmpareProcess_NoMatch(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*compOut) = compOut{BadAccents: []string{"a{x}"}, RC: 1}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Equal(t, utils.NewErrBadAccent([]string{"a{x}"}), err)
}

func TestCpmpareProcess_BadAccents(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	pr, _ := NewComparator("http://server")
	assert.NotNil(t, pr)
	pr.(*comparator).httpWrap = httpJSONMock
	pegomock.When(httpJSONMock.InvokeJSON(pegomock.AnyInterface(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*compOut) = compOut{BadAccents: []string{}, RC: 0}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(d)
	assert.Equal(t, utils.ErrTextDoesNotMatch, err)
}
