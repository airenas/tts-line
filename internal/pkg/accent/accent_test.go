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
