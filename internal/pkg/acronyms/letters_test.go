package acronyms

import (
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/acronyms/service/api"
	"github.com/stretchr/testify/assert"
)

func TestNewWord(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("AAWBA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 3, len(res))
	if len(res) == 3 {
		assert.Equal(t, "aa", res[0].Word)
		assert.Equal(t, "ąą", res[0].WordTrans)
		assert.Equal(t, "ą-ą", res[0].Syll)
		assert.Equal(t, "ąą3", res[0].UserTrans)

		assert.Equal(t, "w", res[1].Word)
		assert.Equal(t, "dablvė", res[1].WordTrans)
		assert.Equal(t, "dabl-vė", res[1].Syll)
		assert.Equal(t, "da4blvė", res[1].UserTrans)

		assert.Equal(t, "ba", res[2].Word)
		assert.Equal(t, "bėą", res[2].WordTrans)
		assert.Equal(t, "bė-ą", res[2].Syll)
		assert.Equal(t, "bėą3", res[2].UserTrans)
	}
}

func TestNewWordFirst(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("WBA", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res))
	if len(res) == 2 {
		assert.Equal(t, "w", res[0].Word)
		assert.Equal(t, "dablvė", res[0].WordTrans)
		assert.Equal(t, "dabl-vė", res[0].Syll)
		assert.Equal(t, "da4blvė", res[0].UserTrans)

		assert.Equal(t, "ba", res[1].Word)
		assert.Equal(t, "bėą", res[1].WordTrans)
		assert.Equal(t, "bė-ą", res[1].Syll)
		assert.Equal(t, "bėą3", res[1].UserTrans)
	}
}

func TestNewWordLast(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("AAW", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, len(res))
	if len(res) == 2 {
		assert.Equal(t, "aa", res[0].Word)
		assert.Equal(t, "ąą", res[0].WordTrans)
		assert.Equal(t, "ą-ą", res[0].Syll)
		assert.Equal(t, "ąą3", res[0].UserTrans)

		assert.Equal(t, "w", res[1].Word)
		assert.Equal(t, "dablvė", res[1].WordTrans)
		assert.Equal(t, "dabl-vė", res[1].Syll)
		assert.Equal(t, "da4blvė", res[1].UserTrans)
	}
}

func TestAddDot(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("A.LT", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	if assert.Equal(t, 3, len(res)) {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)

		assert.Equal(t, "taškas", res[1].Word)
		assert.Equal(t, "taškas", res[1].WordTrans)
		assert.Equal(t, "ta-škas", res[1].Syll)
		assert.Equal(t, "ta3škas", res[1].UserTrans)

		assert.Equal(t, "lt", res[2].Word)
		assert.Equal(t, "eltė", res[2].WordTrans)
		assert.Equal(t, "el-tė", res[2].Syll)
		assert.Equal(t, "eltė3", res[2].UserTrans)
	}
}

func TestNoFail(t *testing.T) {
	prv, err := NewLetters()
	assert.Nil(t, err)
	res, err := prv.Process("'.A.", "1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	if assert.Equal(t, 1, len(res)) {
		assert.Equal(t, "a", res[0].Word)
		assert.Equal(t, "ą", res[0].WordTrans)
		assert.Equal(t, "ą", res[0].Syll)
		assert.Equal(t, "ą3", res[0].UserTrans)
	}
}

func TestLetters_Process(t *testing.T) {
	type args struct {
		word string
		mi   string
	}
	tests := []struct {
		name    string
		s       *Letters
		args    args
		want    []api.ResultWord
		wantErr bool
	}{
		{name: "Simple", args: args{word: "a", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "a", WordTrans: "ą", Syll: "ą", UserTrans: "ą3"}}},
		{name: "To lower", args: args{word: "A", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "a", WordTrans: "ą", Syll: "ą", UserTrans: "ą3"}}},
		{name: "To lower", args: args{word: "B", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "b", WordTrans: "bė", Syll: "bė", UserTrans: "bė3"}}},
		{name: "Ignore dot", args: args{word: "A.", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "a", WordTrans: "ą", Syll: "ą", UserTrans: "ą3"}}},
		{name: "Several", args: args{word: "BA", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "ba", WordTrans: "bėą", Syll: "bė-ą", UserTrans: "bėą3"}}},
		{name: "Accent", args: args{word: "AABA", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "aaba", WordTrans: "ąąbėą", Syll: "ą-ą-bė-ą", UserTrans: "ąąbėą3"}}},
		{name: "Several words", args: args{word: "'.A.W>", mi: "1"}, wantErr: false,
			want: []api.ResultWord{{Word: "a", WordTrans: "ą", Syll: "ą", UserTrans: "ą3"},
				{Word: "w", WordTrans: "dablvė", Syll: "dabl-vė", UserTrans: "da4blvė"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Letters{}
			got, err := s.Process(tt.args.word, tt.args.mi)
			if (err != nil) != tt.wantErr {
				t.Errorf("Letters.Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Letters.Process() = %v, want %v", got, tt.want)
			}
		})
	}
}
