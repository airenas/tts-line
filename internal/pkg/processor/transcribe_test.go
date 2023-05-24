package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTranscriber(t *testing.T) {
	initTestJSON(t)
	pr, err := NewTranscriber("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewTranscriber_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewTranscriber("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeTranscriber(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewTranscriber("http://server")
	assert.NotNil(t, pr)
	pr.(*transcriber).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]transOutput) = []transOutput{{Word: "word",
				Transcription: []trans{{Transcription: "w o r d"}}}}
		}).Return(nil)
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, "w o r d", d.Words[0].Transcription)
}

func TestInvokeTranscriber_SkipFormat(t *testing.T) {
	initTestJSON(t)
	d := newTestTTSDataPart()
	d.Cfg.Input.OutputFormat = api.AudioNone
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	pr, _ := NewTranscriber("http://server")
	pr.(*transcriber).httpWrap = httpJSONMock
	err := pr.Process(d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 0)
}

func TestInvokeTranscriber_FailInput(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewTranscriber("http://server")
	assert.NotNil(t, pr)
	pr.(*transcriber).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: nil})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestInvokeTranscriber_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewTranscriber("http://server")
	assert.NotNil(t, pr)
	pr.(*transcriber).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestInvokeTranscriber_NoData(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewTranscriber("http://server")
	assert.NotNil(t, pr)
	d := newTestTTSDataPart()
	err := pr.Process(d)
	assert.Nil(t, err)
}

func TestInvokeTranscriber_Skip(t *testing.T) {
	initTestJSON(t)
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	pr, _ := NewTranscriber("http://server")
	pr.(*transcriber).httpWrap = httpJSONMock
	err := pr.Process(d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 0)
}

func TestInvokeTranscriber_FailOutput(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewTranscriber("http://server")
	assert.NotNil(t, pr)
	pr.(*transcriber).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]transOutput) = []transOutput{}
		}).Return(nil)
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestMapTransInput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	inp, err := mapTransInput(d)
	assert.Nil(t, err)
	assert.Equal(t, "olia", inp[0].Word)
	assert.Equal(t, 103, inp[0].Acc)
	assert.Equal(t, "o-lia", inp[0].Syll)
	assert.Equal(t, "word", inp[1].Word)
	assert.Equal(t, 103, inp[1].Acc)
	assert.Equal(t, "", inp[1].Syll)
}

func TestMapTransInput_RC(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word2"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103}})
	inp, err := mapTransInput(d)
	require.Nil(t, err)
	require.Equal(t, 4, len(inp))
	assert.Equal(t, "word", inp[0].Rc)
	assert.Equal(t, "word1", inp[1].Rc)
	assert.Equal(t, "", inp[2].Rc)
}

func TestMapTransInput_UserAccent(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia",
		Tagged:        synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"},
		UserAccent:    101})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia",
		Tagged:        synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"}})
	inp, err := mapTransInput(d)
	assert.Nil(t, err)
	assert.Equal(t, 101, inp[0].Acc)
	assert.Equal(t, 103, inp[1].Acc)
}

func TestMapTransInput_CliticsAccent(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia",
		Tagged:        synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"},
		Clitic:        synthesizer.Clitic{Type: synthesizer.CliticsCustom, Accent: 101}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia",
		Tagged:        synthesizer.TaggedWord{Word: "v1"},
		AccentVariant: &synthesizer.AccentVariant{Accent: 103, Syll: "o-lia"},
		Clitic:        synthesizer.Clitic{Type: synthesizer.CliticsNone}})
	inp, err := mapTransInput(d)
	assert.Nil(t, err)
	assert.Equal(t, 101, inp[0].Acc)
	assert.Equal(t, 0, inp[1].Acc)
}

func TestMapTransInput_User(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", UserTranscription: "olia3",
		Tagged: synthesizer.TaggedWord{Word: "v1"},
	})
	inp, err := mapTransInput(d)
	assert.Nil(t, err)
	assert.Equal(t, "olia", inp[0].Word)
	assert.Equal(t, "olia3", inp[0].User)
}

func TestMapTransInput_Space(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", UserTranscription: "olia3",
		Tagged: synthesizer.TaggedWord{Word: "v1"},
	})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	inp, err := mapTransInput(d)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(inp)) {
		assert.Equal(t, "olia", inp[0].Word)
		assert.Equal(t, "olia3", inp[0].User)
	}
}

func TestMapTransOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})

	output := []transOutput{{Word: "olia", Transcription: []trans{{Transcription: "trans"}}},
		{Word: "word", Transcription: []trans{{Transcription: "trans1"}}}}

	err := mapTransOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "trans", d.Words[0].Transcription)
	assert.Equal(t, "trans1", d.Words[1].Transcription)
}

func TestMapTransOutput_Sep(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})

	output := []transOutput{{Word: "olia", Transcription: []trans{{Transcription: "trans"}}},
		{Word: "word", Transcription: []trans{{Transcription: "trans1"}}}}

	err := mapTransOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "trans", d.Words[1].Transcription)
	assert.Equal(t, "trans1", d.Words[3].Transcription)
}

func TestMapTransOutput_DropQMark(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})

	output := []transOutput{{Word: "word", Transcription: []trans{{Transcription: "tran? - s1?"}}}}

	err := mapTransOutput(d, output)
	assert.Nil(t, err)
	assert.Equal(t, "tran - s1", d.Words[0].Transcription)
}

func TestMapTransOutput_FailLen(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})

	output := []transOutput{{Word: "olia", Transcription: []trans{{Transcription: "trans"}}}}

	err := mapTransOutput(d, output)
	assert.NotNil(t, err)
}

func TestMapTransOutput_FailWord(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})

	output := []transOutput{{Word: "olia1", Transcription: []trans{{Transcription: "trans"}}},
		{Word: "word", Transcription: []trans{{Transcription: "trans1"}}}}

	err := mapTransOutput(d, output)
	assert.NotNil(t, err)
}

func TestMapTransOutput_FailError(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{TranscriptionWord: "olia", Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})

	output := []transOutput{{Word: "olia", Transcription: []trans{{Transcription: "trans"}}},
		{Word: "word", Error: "err"}}

	err := mapTransOutput(d, output)
	assert.NotNil(t, err)
}

func Test_addAccent(t *testing.T) {
	type args struct {
		in  string
		acc int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "no acc", args: args{in: "Olia", acc: 0}, want: "Olia"},
		{name: "adds", args: args{in: "Olia", acc: 101}, want: "O4lia"},
		{name: "adds 3", args: args{in: "Olia", acc: 201}, want: "O9lia"},
		{name: "adds 9", args: args{in: "Olia", acc: 301}, want: "O3lia"},
		{name: "with palatalization", args: args{in: "Ol*ia", acc: 301}, want: "O3l*ia"},
		{name: "with palatalization", args: args{in: "Ol*ia", acc: 304}, want: "Ol*ia3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addAccent(tt.args.in, tt.args.acc); got != tt.want {
				t.Errorf("addAccent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserOEPal(t *testing.T) {
	type args struct {
		w *synthesizer.ProcessedWord
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "none", args: args{w: &synthesizer.ProcessedWord{TextPart: &synthesizer.TTSTextPart{Accented: "aa", UserOEPal: ""}}}, want: ""},
		{name: "adds", args: args{w: &synthesizer.ProcessedWord{TextPart: &synthesizer.TTSTextPart{Accented: "aa", UserOEPal: "aa"}}}, want: "aa"},
		{name: "with acc", args: args{w: &synthesizer.ProcessedWord{UserAccent: 302, TextPart: &synthesizer.TTSTextPart{Accented: "aa", UserOEPal: "aa"}}}, want: "aa3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUserOEPal(tt.args.w); got != tt.want {
				t.Errorf("getUserOEPal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSyllables(t *testing.T) {
	type args struct {
		w *synthesizer.ProcessedWord
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "none", args: args{w: &synthesizer.ProcessedWord{TextPart: &synthesizer.TTSTextPart{Accented: "aa", Syllables: ""}}}, want: ""},
		{name: "adds", args: args{w: &synthesizer.ProcessedWord{TextPart: &synthesizer.TTSTextPart{Accented: "aa", Syllables: "a-a"}}}, want: "a-a"},
		{name: "from accenter", args: args{w: &synthesizer.ProcessedWord{UserAccent: 302,
			TextPart: &synthesizer.TTSTextPart{Accented: "aa"}, AccentVariant: &synthesizer.AccentVariant{Syll: "a-b"}}}, want: "a-b"},
		{name: "prefers user's", args: args{w: &synthesizer.ProcessedWord{UserAccent: 302,
			TextPart: &synthesizer.TTSTextPart{Accented: "aa", Syllables: "a-a"}, AccentVariant: &synthesizer.AccentVariant{Syll: "a-b"}}}, want: "a-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSyllables(tt.args.w); got != tt.want {
				t.Errorf("getSyllables() = %v, want %v", got, tt.want)
			}
		})
	}
}
