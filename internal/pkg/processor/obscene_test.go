package processor

import (
	"context"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewObsceneFilter(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{name: "OK", args: "http://server:8000", wantErr: false},
		{name: "Fail", args: "http://", wantErr: true},
		{name: "Fail", args: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewObsceneFilter(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewObsceneFilter() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestInvokeObscene(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewObsceneFilter("http://server")
	assert.NotNil(t, pr)
	pr.(*obscene).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]obsceneResultToken) = []obsceneResultToken{{Token: "word", Obscene: 1}}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, true, d.Words[0].Obscene)
}

func TestInvokeObscene_FailOutput(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewObsceneFilter("http://server")
	assert.NotNil(t, pr)
	pr.(*obscene).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]obsceneResultToken) = []obsceneResultToken{}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestInvokeObscene_Skip(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewObsceneFilter("http://server")
	assert.NotNil(t, pr)
	pr.(*obscene).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Cfg.JustAM = true
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 0)
}

func TestInvokeObscene_FailError(t *testing.T) {
	initTestJSON(t)
	pr, _ := NewObsceneFilter("http://server")
	assert.NotNil(t, pr)
	pr.(*obscene).httpWrap = httpJSONMock
	d := newTestTTSDataPart()
	d.Words = append(d.Words, &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "word"}})
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("err"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func Test_mapObsceneInput(t *testing.T) {
	tests := []struct {
		name string
		args []*synthesizer.ProcessedWord
		want []*obsceneToken
	}{
		{name: "One word", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}}},
			want: []*obsceneToken{{Token: "word"}}},
		{name: "Several word", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{Word: "word2"}}},
			want: []*obsceneToken{{Token: "word"}, {Token: "word2"}}},
		{name: "Skip sentence", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{Word: "word2"}}, {Tagged: synthesizer.TaggedWord{SentenceEnd: true}}},
			want: []*obsceneToken{{Token: "word"}, {Token: "word2"}}},
		{name: "Skip sep", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{Separator: ","}},
			{Tagged: synthesizer.TaggedWord{Word: "word2"}}},
			want: []*obsceneToken{{Token: "word"}, {Token: "word2"}}},
		{name: "User transcription", args: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"},
			UserTranscription: "olia"}},
			want: []*obsceneToken{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestTTSDataPart()
			d.Words = append(d.Words, tt.args...)
			if got := mapObsceneInput(d); !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_mapObsceneOutput(t *testing.T) {
	type args struct {
		wrds []*synthesizer.ProcessedWord
		out  []obsceneResultToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []bool
	}{
		{name: "One word", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}}},
			out: []obsceneResultToken{{Token: "word", Obscene: 0}}},
			want: []bool{false}, wantErr: false},
		{name: "One word set", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}}},
			out: []obsceneResultToken{{Token: "word", Obscene: 1}}},
			want: []bool{true}, wantErr: false},
		{name: "Several set", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{Word: "word2"}}},
			out: []obsceneResultToken{{Token: "word", Obscene: 1}, {Token: "word2", Obscene: 1}}},
			want: []bool{true, true}, wantErr: false},
		{name: "Sentence", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{SentenceEnd: true}}},
			out: []obsceneResultToken{{Token: "word", Obscene: 1}}},
			want: []bool{true}, wantErr: false},
		{name: "Fail", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}},
			{Tagged: synthesizer.TaggedWord{Word: "word2"}}},
			out: []obsceneResultToken{{Token: "word", Obscene: 1}}},
			want: []bool{true, false}, wantErr: true},
		{name: "Fail word", args: args{wrds: []*synthesizer.ProcessedWord{{Tagged: synthesizer.TaggedWord{Word: "word"}}},
			out: []obsceneResultToken{{Token: "wordx", Obscene: 1}}},
			want: []bool{false}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestTTSDataPart()
			d.Words = append(d.Words, tt.args.wrds...)
			if err := mapObsceneOutput(d, tt.args.out); (err != nil) != tt.wantErr {
				t.Errorf("mapObsceneOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
			for i, b := range tt.want {
				assert.Equal(t, b, d.Words[i].Obscene, "wrong obscene for %s", d.Words[i].Tagged.Word)
			}
		})
	}
}
