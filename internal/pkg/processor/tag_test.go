package processor

import (
	"errors"
	"testing"

	"github.com/petergtz/pegomock"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
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

func TestCreateTaggerAccent(t *testing.T) {
	initTest(t)
	pr, err := NewTaggerAccents("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateTaggerAccent_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewTaggerAccents("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeTaggerAccent(t *testing.T) {
	initTest(t)
	pr, _ := NewTaggerAccents("http://server")
	assert.NotNil(t, pr)
	pr.(*taggerAccents).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	d.OriginalText = " ,word"
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

func TestInvokeTaggerAccent_Fail(t *testing.T) {
	initTest(t)
	pr, _ := NewTaggerAccents("http://server")
	assert.NotNil(t, pr)
	pr.(*taggerAccents).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestInvokeTaggerAccent_FailMap(t *testing.T) {
	initTest(t)
	pr, _ := NewTaggerAccents("http://server")
	assert.NotNil(t, pr)
	pr.(*taggerAccents).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	d.OriginalText = " ,wor1"
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*[]*TaggedWord) = []*TaggedWord{{Type: "SPACE", String: " "},
				{Type: "SEPARATOR", String: ","}, {Type: "WORD", String: "word", Lemma: "lemma", Mi: "mi"},
				{Type: "SENTENCE_END"}}
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestClearAccents(t *testing.T) {
	tests := []struct {
		v string
		e string
	}{
		{v: "", e: ""},
		{v: "{a~}", e: "a"},
		{v: "{a\\}", e: "a"},
		{v: "{a/}", e: "a"},
		{v: "{a~", e: "{a~"},
		{v: "{a~ }", e: "{a~ }"},
		{v: "{1~}", e: "{1~}"},
		{v: "{Ą~}", e: "Ą"},
		{v: "{ą~}", e: "ą"},
		{v: "oli{ą~} {ą~}s", e: "olią ąs"},
		{v: "oli{ą~}{k~} {{ą~}}s", e: "oliąk {ą}s"},
	}

	for i, tc := range tests {
		v := clearAccents(tc.v)
		assert.Equal(t, tc.e, v, "Fail %d", i)
	}
}

func TestMapAccent(t *testing.T) {
	p, err := mapTagAccentResult([]*TaggedWord{{Type: "SEPARATOR", String: ","},
		{Type: "SENTENCE_END"}}, ",")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(p)) {
		assert.Equal(t, ",", p[0].Tagged.Separator)
		assert.Equal(t, true, p[1].Tagged.SentenceEnd)
	}

	p, err = mapTagAccentResult([]*TaggedWord{{Type: "WORD", String: "mama"},
		{Type: "SPACE", String: " "}}, "mam{a~} ")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(p)) {
		assert.Equal(t, "mama", p[0].Tagged.Word)
		assert.Equal(t, 304, p[0].UserAccent)
		assert.Equal(t, true, p[1].Tagged.Space)
	}

	p, err = mapTagAccentResult([]*TaggedWord{{Type: "SPACE", String: " "},
		{Type: "WORD", String: "mama"}}, " mam{a~}")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(p)) {
		assert.Equal(t, true, p[0].Tagged.Space)
		assert.Equal(t, "mama", p[1].Tagged.Word)
		assert.Equal(t, 304, p[1].UserAccent)
	}

	p, err = mapTagAccentResult([]*TaggedWord{
		{Type: "WORD", String: "mama"}, {Type: "WORD", String: "mama"}}, "mam{a~}mam{a~}")
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(p)) {
		assert.Equal(t, "mama", p[0].Tagged.Word)
		assert.Equal(t, 304, p[0].UserAccent)
		assert.Equal(t, "mama", p[1].Tagged.Word)
		assert.Equal(t, 304, p[1].UserAccent)
	}
}

func TestMapAccent_Fail(t *testing.T) {
	_, err := mapTagAccentResult([]*TaggedWord{{Type: "WORD", String: "mama"}}, " mam{a~}")
	assert.NotNil(t, err)
	_, err = mapTagAccentResult([]*TaggedWord{{Type: "WORD", String: "mama"}}, ",mam{a~}")
	assert.NotNil(t, err)
}

func TestMapAccent_ErrorType(t *testing.T) {
	_, err := mapTagAccentResult([]*TaggedWord{{Type: "WORD", String: "mama"}}, "m{a~}m{a~}")
	assert.NotNil(t, err)
	var errBA *utils.ErrBadAccent
	assert.True(t, errors.As(err, &errBA))

	_, err = mapTagAccentResult([]*TaggedWord{{Type: "WORD", String: "mama"}}, ",mam{a~}")
	assert.NotNil(t, err)
	assert.False(t, errors.As(err, &errBA))
}

func TestMapTag(t *testing.T) {
	tests := []struct {
		v TaggedWord
		e synthesizer.TaggedWord
	}{
		{v: TaggedWord{Type: "WORD", String: "mama", Mi: "mi"}, e: synthesizer.TaggedWord{Word: "mama", Mi: "mi"}},
		{v: TaggedWord{Type: "NUMBER", String: "10", Mi: "mi"}, e: synthesizer.TaggedWord{Word: "10", Mi: "mi"}},
		{v: TaggedWord{Type: "SPACE", String: "  "}, e: synthesizer.TaggedWord{Space: true}},
		{v: TaggedWord{Type: "SEPARATOR", String: ","}, e: synthesizer.TaggedWord{Separator: ","}},
	}

	for _, tc := range tests {
		v := mapTag(&tc.v)
		assert.Equal(t, tc.e, v)
	}
}
