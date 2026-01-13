package processor

import (
	"context"
	"reflect"
	"testing"

	aapi "github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	httpJSONMock *mocks.HTTPInvokerJSON
)

func initTestJSON(t *testing.T) {
	t.Helper()

	httpJSONMock = &mocks.HTTPInvokerJSON{}
}

func TestNewAbbreviator(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcronyms("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewAbbreviator_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewAcronyms("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeNewAbbreviator(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcronyms("http://server")
	assert.NotNil(t, pr)
	pr.(*acronyms).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Mi: "Y"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]acrWordOutput) = []acrWordOutput{{ID: "0",
				Words: []acrResultWord{{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}}}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "o l i a", d.Words[0].UserTranscription)
	assert.Equal(t, "o-lia", d.Words[0].UserSyllables)
	assert.Equal(t, "olia", d.Words[0].Tagged.Word)
}

func TestInvokeNewAbbreviator_Fail(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewAcronyms("http://server")
	assert.NotNil(t, pr)
	pr.(*acronyms).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Mi: "Y"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}
func TestInvokeAbbr_Skip(t *testing.T) {
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	pr, _ := NewAcronyms("http://server")
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
}

func TestMapAbbrOutput(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Space: true}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "0", Words: []acrResultWord{
		{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.Nil(t, err)
	assert.Equal(t, ",", d.Words[2].Tagged.Separator)
	assert.Equal(t, "o l i a", d.Words[0].UserTranscription)
	assert.Equal(t, "olia", d.Words[0].Tagged.Word)
}

func TestMapAbbrOutput_Several(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}})
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "1", Words: []acrResultWord{
		{Word: "v1", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"},
		{Word: "v2", WordTrans: "oolia", UserTrans: "v 2", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.Nil(t, err)
	assert.Equal(t, ",", d.Words[0].Tagged.Separator)
	assert.Equal(t, "v1", d.Words[1].Tagged.Word)
	assert.Equal(t, "v2", d.Words[2].Tagged.Word)
}

func TestMapAbbrOutput_Fail(t *testing.T) {
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "v1"}})

	abbrOut := []acrWordOutput{}
	abbrOut = append(abbrOut, acrWordOutput{ID: "XX", Words: []acrResultWord{
		{Word: "olia", WordTrans: "oolia", UserTrans: "o l i a", Syll: "o-lia"}}})

	err := mapAbbrOutput(d, abbrOut)
	assert.NotNil(t, err)
}

func TestIsAbbr(t *testing.T) {
	assert.True(t, isAbbr("Ya", "O"))
	assert.True(t, isAbbr("X", "O"))
	assert.True(t, isAbbr("N", "PP"))
	assert.False(t, isAbbr("Npmsnng", "Kaunas"))
	assert.False(t, isAbbr("Npmsnng", "Pp"))
}

func newTestTTSDataPart() *synthesizer.TTSDataPart {
	return &synthesizer.TTSDataPart{Cfg: &synthesizer.TTSConfig{Input: &api.TTSRequestConfig{OutputFormat: api.AudioMP3}}}
}

func Test_mapAbbrInput(t *testing.T) {
	tests := []struct {
		name string
		args []*synthesizer.ProcessedWord
		want []aapi.WordInput
	}{
		{name: "One word", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "Xolia"}}},
			want: []aapi.WordInput{{Word: "word", MI: "Xolia", ID: "0"}}},
		{name: "No mi", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: ""}}},
			want: []aapi.WordInput{}},
		{name: "Obscene", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: ""}, Obscene: true}},
			want: []aapi.WordInput{{Word: "word", MI: "", ID: "0", Mode: aapi.ModeCharactersAsWord}}},
		{name: "AllUpper", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "Naaa",
			Lemma: "WORD"}, Obscene: false}},
			want: []aapi.WordInput{{Word: "word", MI: "Naaa", ID: "0"}}},
		{name: "As chars", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "Naaa",
			Lemma: "WORD"}, Obscene: false, TextPart: &synthesizer.TTSTextPart{InterpretAs: ssml.InterpretAsTypeCharacters}}},
			want: []aapi.WordInput{{Word: "word", MI: "Naaa", ID: "0", Mode: aapi.ModeCharacters}}},
		{name: "As chars details", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "Naaa",
			Lemma: "WORD"}, Obscene: false, TextPart: &synthesizer.TTSTextPart{InterpretAs: ssml.InterpretAsTypeCharacters,
			InterpretAsDetail: ssml.InterpretAsDetailTypeReadSymbols}}},
			want: []aapi.WordInput{{Word: "word", MI: "Naaa", ID: "0", Mode: aapi.ModeAllAsCharacters}}},
		{name: "No AllUpper", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "Naaa",
			Lemma: "WORd"}, Obscene: false}},
			want: []aapi.WordInput{}},
		{name: "Skip user accent", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "X-",
			Lemma: "Looong"}, UserAccent: 101}},
			want: []aapi.WordInput{}},
		{name: "Obscene with user accent",
			args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word", Mi: "X-", Lemma: "Looooong"},
				Obscene: true, UserAccent: 101}},
			want: []aapi.WordInput{{Word: "word", MI: "X-", ID: "0", Mode: aapi.ModeCharactersAsWord}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestTTSDataPart()
			d.Words = append(d.Words, tt.args...)
			if got := mapAbbrInput(d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapAbbrInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAccented(t *testing.T) {
	type args struct {
		w *synthesizer.ProcessedWord
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "accented", args: args{w: &synthesizer.ProcessedWord{UserAccent: 101}}, want: true},
		{name: "accented", args: args{w: &synthesizer.ProcessedWord{UserAccent: 0, TextPart: &synthesizer.TTSTextPart{Accented: "olia"}}}, want: true},
		{name: "not accented", args: args{w: &synthesizer.ProcessedWord{UserAccent: 0, TextPart: &synthesizer.TTSTextPart{Accented: ""}}}, want: false},
		{name: "not accented", args: args{w: &synthesizer.ProcessedWord{UserAccent: 0}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAccented(tt.args.w); got != tt.want {
				t.Errorf("isAccented() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasUserTranscriptions(t *testing.T) {
	type args struct {
		w *synthesizer.ProcessedWord
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "transcription", args: args{w: &synthesizer.ProcessedWord{UserTranscription: "olia"}}, want: true},
		{name: "syllables", args: args{w: &synthesizer.ProcessedWord{UserSyllables: "olia"}}, want: true},
		{name: "accent", args: args{w: &synthesizer.ProcessedWord{UserAccent: 101}}, want: true},
		{name: "empty", args: args{w: &synthesizer.ProcessedWord{UserAccent: 0}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasUserTranscriptions(tt.args.w); got != tt.want {
				t.Errorf("hasUserTranscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_needAbbrProcessing(t *testing.T) {
	tests := []struct {
		name string
		w    *synthesizer.ProcessedWord
		want bool
	}{
		{name: "not a word", w: &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Separator: ","}}, want: false},
		{name: "regular word", w: &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "olia"}}, want: false},
		{name: "one letter", w: &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "o"}, NERType: synthesizer.NERSingleLetter}, want: true},
		{name: "greek", w: &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "o"}, NERType: synthesizer.NERGreekLetters}, want: true},
		{name: "Accented", w: &synthesizer.ProcessedWord{UserAccent: 301, Tagged: synthesizer.TaggedWord{Word: "o"}, NERType: synthesizer.NERSingleLetter}, want: false},
		{name: "Obscene", w: &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "o"}, NERType: synthesizer.NERRegular, Obscene: true}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := needAbbrProcessing(tt.w)
			if got != tt.want {
				t.Errorf("needAbbrProcessing() = %v, want %v", got, tt.want)
			}
		})
	}
}
