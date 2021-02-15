package processor

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
)

func TestCreateTagger(t *testing.T) {
	initTest(t)
	pr, err := NewTagger("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateTagger_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewTagger("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeTagger(t *testing.T) {
	initTest(t)
	pr, _ := NewTagger("http://server")
	assert.NotNil(t, pr)
	pr.(*tagger).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]*TaggedWord) = []*TaggedWord{{Type: "SPACE", String: " "},
				{Type: "SEPARATOR", String: ","}, {Type: "WORD", String: "word", Lemma: "lemma", Mi: "mi"},
				{Type: "SENTENCE_END"}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(d.Words))
	assert.Equal(t, true, d.Words[0].Tagged.Space)
	assert.False(t, d.Words[0].Tagged.IsWord())

	assert.Equal(t, ",", d.Words[1].Tagged.Separator)
	assert.Equal(t, "", d.Words[1].Tagged.Word)
	assert.False(t, d.Words[1].Tagged.IsWord())

	assert.Equal(t, "", d.Words[2].Tagged.Separator)
	assert.Equal(t, "word", d.Words[2].Tagged.Word)
	assert.Equal(t, "lemma", d.Words[2].Tagged.Lemma)
	assert.Equal(t, "mi", d.Words[2].Tagged.Mi)
	assert.True(t, d.Words[2].Tagged.IsWord())

	assert.True(t, d.Words[3].Tagged.SentenceEnd)
	assert.False(t, d.Words[3].Tagged.IsWord())
}

func TestInvokeTagger_Fail(t *testing.T) {
	initTest(t)
	pr, _ := NewTagger("http://server")
	assert.NotNil(t, pr)
	pr.(*tagger).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestInvokeTagger_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewTagger("http://server")
	err := pr.Process(d)
	assert.Nil(t, err)
}
