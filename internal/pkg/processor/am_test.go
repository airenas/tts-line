package processor

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test"
	"github.com/airenas/tts-line/pkg/ssml"
)

func TestNewAcousticModel(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewAcousticModel_Space(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.NotNil(t, pr)
	assert.Equal(t, "sil", pr.(*amodel).spaceSymbol)
	assert.Equal(t, "sil", pr.(*amodel).endSymbol)
	pr, _ = NewAcousticModel(test.NewConfig(t, "url: http://server\nspaceSymbol: <space>"))
	assert.NotNil(t, pr)
	assert.Equal(t, "<space>", pr.(*amodel).spaceSymbol)
	assert.Equal(t, "<space>", pr.(*amodel).endSymbol)
}

func TestNewAcousticModel_EndSymbol(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\nendSymbol: <end>"))
	assert.NotNil(t, pr)
	assert.Equal(t, "<end>", pr.(*amodel).endSymbol)
}

func TestNewAcousticModel_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcousticModel(nil)
	assert.NotNil(t, err)
	assert.Nil(t, pr)
	pr, err = NewAcousticModel(test.NewConfig(t, ""))
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestNewAcousticModel_ReadVocoder(t *testing.T) {
	initTestJSON(t)
	pr := newTestAMCfg(t, test.NewConfig(t, "url: http://server\nhasVocoder: true"))
	assert.True(t, pr.hasVocoder)
	pr = newTestAMCfg(t, test.NewConfig(t, "url: http://server"))
	assert.False(t, pr.hasVocoder)
}

func TestInvokeAcousticModel(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://{{voice}}.server\n"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Spectogram = "spectogram"
	d.Cfg.Prosodies = []*ssml.Prosody{{Rate: 0.5}, {Rate: 1.0}}
	d.Cfg.Voice = "aa"
	d.Cfg.Input.Priority = 10
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[2].(*amOutput) = amOutput{Data: "spec"}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "spec", d.Spectogram)

	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSONU", 1)
	url := httpJSONMock.Calls[0].Arguments[0]
	inp := httpJSONMock.Calls[0].Arguments[1]

	ai := inp.(*amInput)
	assert.InDelta(t, 0.5, ai.Speed, 0.0001)
	assert.Equal(t, "aa", ai.Voice)
	assert.Equal(t, 10, ai.Priority)
	assert.Equal(t, "http://aa.server", url)
}

func TestInvokeAcousticModel_Skip(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Cfg.Input.OutputFormat = api.AudioNone
	d.Spectogram = ""
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSONU", 0)
	assert.Equal(t, d.TranscribedText, "")
}

func TestInvokeAcousticModel_Skip_ReturnTranscribed(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Cfg.Input.OutputFormat = api.AudioNone
	d.Cfg.Input.OutputTextFormat = api.TextTranscribed
	d.Spectogram = ""
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, d.TranscribedText, "sil")
}

func TestInvokeAcousticModel_WriteAudio(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\nhasVocoder: true"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Spectogram = "spectogram"
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[2].(*amOutput) = amOutput{Data: "audio"}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "spectogram", d.Spectogram)
	assert.Equal(t, "audio", d.Audio)
	assert.Equal(t, d.TranscribedText, "sil")
}

func TestInvokeAcousticModel_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Spectogram = "spectogram"
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestInvokeAcousticModel_FromAM(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcousticModel(test.NewConfig(t, "url: http://server\n"))
	assert.NotNil(t, pr)
	pr.(*amodel).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	d.Text = "olia"
	httpJSONMock.On("InvokeJSONU", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSONU", 1)
	cp1 := httpJSONMock.Calls[0].Arguments[1]
	assert.Equal(t, &amInput{Text: "olia", Speed: 1}, cp1)
}

func TestMapAMInput(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a o l i a , v a <space>", inp.Text)
}

// func TestMapAMInput_NoSilAtStart(t *testing.T) {
// 	pr := newTestAM(t, "http://server", "sil")
// 	d := newTestTTSDataPart()
// 	d.First = false

// 	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
// 	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
// 	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
// 	inp := pr.mapAMInput(d)
// 	assert.Equal(t, "v a o l i a , v a sil", inp.Text)
// }

func TestMapAMInput_SpaceDot(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a o l i a . <space> v a <space>", inp.Text)
}

func TestMapAMInput_SpaceQuestion(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a o l i a ? <space> v a <space>", inp.Text)
}

func TestMapAMInput_Exclamation(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true

	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - o l i a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "!"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a o l i a ! <space> v a <space>", inp.Text)
}

func TestMapAMInput_ChangePhone(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()

	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a - \"oi \"iui", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a \"o: i \"iu i <space>", inp.Text)
}

func TestMapAMInput_SpaceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true

	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a <space>", inp.Text)
}

func TestMapAMInput_AddDotOnSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sil", inp.Text)
}

func TestMapAMInput_NoDotOnSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sil", inp.Text)
}

func TestMapAMInput_NoDotOnSentenceEnd2(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a ? sil", inp.Text)
}

func TestMapAMInput_SeveralSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "?"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sil v a ? sil", inp.Text)
}

func TestMapAMInput_CustomEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	pr.endSymbol = "sp sil"
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sp sil", inp.Text)
}

func TestMapAMInput_CustomEnd_Dot(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	pr.endSymbol = "sp sil"
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sp sil", inp.Text)
}

func TestMapAMInput_CustomEnd_SeveralSentenceEnd(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	pr.endSymbol = "sp sil"
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sil v a . sp sil", inp.Text)
}

func TestMapAMInput_DropDash(t *testing.T) {
	pr := newTestAM(t, "http://server", "<space>")
	d := newTestTTSDataPart()
	d.First = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "<space> v a v a - <space> v a <space>", inp.Text)
}

func TestMapAMInput_SkipSpaces(t *testing.T) {
	pr := newTestAM(t, "http://server", "sil")
	pr.endSymbol = "sp sil"
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Transcription: "v a", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{SentenceEnd: true}})
	inp, _ := pr.mapAMInput(d)
	assert.Equal(t, "sil v a . sil v a v a . sp sil", inp.Text)
}

func TestAddPause(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "."}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	assert.True(t, addPause(".", d.Words, 1))
	assert.True(t, addPause("?", d.Words, 1))
	assert.True(t, addPause("!", d.Words, 1))
	assert.True(t, addPause("-", d.Words, 1))
	assert.True(t, addPause(":", d.Words, 1))

	d = newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	assert.False(t, addPause("-", d.Words, 1))
	assert.False(t, addPause(":", d.Words, 1))
	assert.True(t, addPause(".", d.Words, 1))
	assert.True(t, addPause("?", d.Words, 1))
	assert.True(t, addPause("!", d.Words, 1))

	d = newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	assert.True(t, addPause("-", d.Words, 1))
	assert.True(t, addPause(":", d.Words, 1))

	d = newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	assert.True(t, addPause("-", d.Words, 0))
	assert.True(t, addPause(":", d.Words, 0))
}

func TestSep(t *testing.T) {
	tw := make([]*synthesizer.ProcessedWord, 0)
	tw = append(tw, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	for _, s := range ",:.?!-" {
		assert.Equal(t, string(s), getSep(string(s), tw, 0))
	}
	assert.Equal(t, "...", getSep("...", tw, 0))
	assert.Equal(t, ",", getSep(";", tw, 0))
	assert.Equal(t, "", getSep("\"", tw, 0))
	tw = make([]*synthesizer.ProcessedWord, 0)
	tw = append(tw, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "aa"}})
	tw = append(tw, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: "-"}})
	tw = append(tw, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "aa"}})
	for _, s := range ":-" {
		assert.Equal(t, "", getSep(string(s), tw, 1))
	}
}

func TestChangePhn(t *testing.T) {
	tests := []struct {
		v string
		e string
	}{
		{v: "-", e: ""},
		{v: "e", e: "e"},
		{v: "\"Eu", e: "eu"},
		{v: "Ou", e: "o u"},
		{v: "\"ou", e: "\"o: u"},
		{v: "^Oi", e: "\"o i"},
		{v: "\"iui", e: "\"iu i"},
		{v: "\"oi", e: "\"o: i"},
		{v: "\"iOi", e: "\"io i"},
		{v: "\"ioi", e: "\"io: i"},
		{v: "iOi", e: "io i"},
		{v: "ioi", e: "io: i"},
	}

	for i, tc := range tests {
		assert.Equal(t, tc.e, changePhn(tc.v), "Fail %d", i)
	}
}

func TestVoiceURL(t *testing.T) {
	tests := []struct {
		u string
		v string
		e string
	}{
		{u: "http://server", v: "voice", e: "http://server"},
		{u: "http://server.{{voice}}", v: "voice", e: "http://server.voice"},
		{u: "http://{{voice}}-server.{{voice}}", v: "voice", e: "http://voice-server.voice"},
	}

	for i, tc := range tests {
		assert.Equal(t, tc.e, getVoiceURL(tc.u, tc.v), "Fail %d", i)
	}
}

func newTestAM(t *testing.T, urlStr string, spaceSym string) *amodel {
	return newTestAMCfg(t, test.NewConfig(t, fmt.Sprintf("url: %s\nspaceSymbol: %s\nendSymbol: %s", urlStr, spaceSym, "")))
}

func newTestAMCfg(t *testing.T, cfg *viper.Viper) *amodel {
	pr, err := NewAcousticModel(cfg)
	assert.Nil(t, err)
	return pr.(*amodel)
}
