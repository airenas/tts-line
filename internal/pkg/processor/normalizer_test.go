package processor

import (
	"context"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewNormalizer(t *testing.T) {
	initTestJSON(t)
	pr, err := NewNormalizer("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewNormalizer_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewNormalizer("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestNormalizeProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.CleanedText = []string{" a a"}
	d.OriginalTextParts = []*synthesizer.TTSTextPart{{Text: " a a"}}

	pr, _ := NewNormalizer("http://server")
	assert.NotNil(t, pr)
	pr.(*normalizer).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*[]*normResponseData) = []*normResponseData{{Res: "out text", Rep: []*normResponseDataRep{
				{Beg: 0, End: 4, Text: "out text"}}}}
		}).Return(nil)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp1 := httpJSONMock.Calls[0].Arguments[0]
	cp1, ok := cp1.(*normRequestData)
	require.True(t, ok)
	assert.Equal(t, &normRequestData{Items: []*normRequestOrigData{{Orig: []*normText{{Text: " a a ", Type: normTextTypePlain}}}}}, cp1)
	assert.Equal(t, []string{"out text"}, d.NormalizedText)
}

func TestNormalizeProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.CleanedText = []string{" a a"}
	d.OriginalTextParts = []*synthesizer.TTSTextPart{{Text: " a a"}}
	pr, _ := NewNormalizer("http://server")
	assert.NotNil(t, pr)
	pr.(*normalizer).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))
	err := pr.Process(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestNormalize_Skip(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewNormalizer("http://server")
	pr.(*normalizer).httpWrap = httpJSONMock
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 0)
}

func Test_processNormalizedOutput(t *testing.T) {
	type args struct {
		output normResponseData
		input  []string
		fixed  []bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "replace several longer", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 19, Text: "aa1010aa"}, {Beg: 35, End: 39, Text: "bbbbbb"}}},
			input: []string{"vienas du olia oooo", "oooo", "vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa1010aa", "oooo", "vienas du bbbbbb oooo"}},
		{name: "replace several longer", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa1010aa"}, {Beg: 35, End: 39, Text: "bbbbbb"}}},
			input: []string{"vienas du olia oooo", "oooo", "vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa1010aa oooo", "oooo", "vienas du bbbbbb oooo"}},
		{name: "fixed", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa1010aa"}, {Beg: 35, End: 39, Text: "bbbbbb"}}},
			fixed: []bool{false, false, true},
			input: []string{"vienas du olia oooo", "oooo", "vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa1010aa oooo", "oooo", "vienas du olia oooo"}},
		{name: "replace several", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa"}, {Beg: 35, End: 39, Text: "bbbbbb"}}},
			input: []string{"vienas du olia oooo", "oooo", "vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa oooo", "oooo", "vienas du bbbbbb oooo"}},
		{name: "replace several longer", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa1010aa"},
			{Beg: 35, End: 39, Text: "bbbbbb"}}},
			input: []string{"vienas du olia oooo", "oooo", "vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa1010aa oooo", "oooo", "vienas du bbbbbb oooo"}},
		{name: "replace several in one", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 5, End: 6, Text: "aa"}, {Beg: 10, End: 12, Text: "bbb"}}},
			input: []string{"01234567890123456789 ", "oooo"}},
			wantErr: false, want: []string{"01234aa6789bbb23456789 ", "oooo"}},
		{name: "replace one", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa"}}, Res: "vienas du aa oooo"}, input: []string{"vienas du olia oooo"}},
			wantErr: false, want: []string{"vienas du aa oooo"}},
		{name: "replace several", args: args{output: normResponseData{Rep: []*normResponseDataRep{{Beg: 10, End: 14, Text: "aa"}}},
			input: []string{"vienas du olia oooo", "oooo"}},
			wantErr: false, want: []string{"vienas du aa oooo", "oooo"}},
		{name: "empty", args: args{output: normResponseData{Rep: []*normResponseDataRep{}}, input: []string{""}}, wantErr: false, want: []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &synthesizer.TTSData{}
			d.CleanedText = tt.args.input
			d.OriginalTextParts = make([]*synthesizer.TTSTextPart, len(tt.args.input))
			for i, v := range tt.args.input {
				d.OriginalTextParts[i] = &synthesizer.TTSTextPart{Text: v}
				if len(tt.args.fixed) > i {
					if tt.args.fixed[i] {
						d.OriginalTextParts[i].InterpretAs = ssml.InterpretAsTypeCharacters
					}
				}
			}
			got, err := processNormalizedOutput(&tt.args.output, d)
			if (err != nil) != tt.wantErr {
				t.Errorf("processNormalizedOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processNormalizedOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isFixedText(t *testing.T) {
	tests := []struct {
		name string
		part *synthesizer.TTSTextPart
		want bool
	}{
		{name: "interpretAs", part: &synthesizer.TTSTextPart{InterpretAs: ssml.InterpretAsTypeCharacters}, want: true},
		{name: "accented", part: &synthesizer.TTSTextPart{Accented: "olia"}, want: true},
		{name: "userOEPal", part: &synthesizer.TTSTextPart{UserOEPal: "olia"}, want: true},
		{name: "none", part: &synthesizer.TTSTextPart{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFixedText(tt.part)
			if got != tt.want {
				t.Errorf("isFixedText() = %v, want %v", got, tt.want)
			}
		})
	}
}
