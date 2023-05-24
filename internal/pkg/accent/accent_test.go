package accent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAccentString(t *testing.T) {
	tests := []struct {
		v   string
		a   int
		e   string
		err bool
	}{
		{v: "mama", a: 102, e: "m{a\\}ma", err: false},
		{v: "mama", a: 202, e: "m{a/}ma", err: false},
		{v: "mama", a: 302, e: "m{a~}ma", err: false},
		{v: "mama", a: 301, e: "{m~}ama", err: false},
		{v: "mama", a: 304, e: "mam{a~}", err: false},
		{v: "a", a: 301, e: "{a~}", err: false},
		{v: "mama", a: 402, e: "", err: true},
		{v: "mama", a: 105, e: "", err: true},
		{v: "mama", a: 0, e: "mama", err: false},
		{v: "ūkūs", a: 303, e: "ūk{ū~}s", err: false},
	}

	for i, tc := range tests {
		v, err := ToAccentString(tc.v, tc.a)
		assert.Equal(t, tc.err, err != nil, "Fail %d", i)
		assert.Equal(t, tc.e, v, "Fail %d", i)
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		v rune
		e int
	}{
		{v: '\\', e: 1},
		{v: '/', e: 2},
		{v: '~', e: 3},
		{v: 'a', e: 0},
	}

	for i, tc := range tests {
		assert.Equal(t, tc.e, Value(tc.v), "Fail %d", i)
	}
}

func TestIsWordOrWithAccent(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{name: "empty", args: "", want: false},
		{name: "word", args: "a", want: true},
		{name: "word", args: "aba", want: true},
		{name: "acc", args: "ab{a/}", want: true},
		{name: "one letter", args: "{a/}", want: true},
		{name: "no Word", args: ",", want: false},
		{name: "wrong", args: "a123", want: false},
		{name: "wrong", args: "a,", want: false},
		{name: "wrong", args: "a b", want: false},
		{name: "wrong", args: "a-c", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWordOrWithAccent(tt.args); got != tt.want {
				t.Errorf("IsWordOrWithAccent() = %v, want %v", got, tt.want)
			}
		})
	}
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
		t.Run(tc.v, func(t *testing.T) {
			v := ClearAccents(tc.v)
			assert.Equal(t, tc.e, v, "Fail %d", i)
		})
	}
}

func TestTranscriberAccent(t *testing.T) {
	type args struct {
		acc int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "none", args:  args{acc: 0}, want: ""},
		{name: "3", args:  args{acc: 3}, want: "3"},
		{name: "9", args:  args{acc: 2}, want: "9"},
		{name: "4", args:  args{acc: 1}, want: "4"},
		{name: "other", args:  args{acc: 10}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranscriberAccent(tt.args.acc); got != tt.want {
				t.Errorf("TranscriberAccent() = %v, want %v", got, tt.want)
			}
		})
	}
}
