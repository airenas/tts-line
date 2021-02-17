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
		{v: "mama", a: 102, e: "ma{\\}ma", err: false},
		{v: "mama", a: 202, e: "ma{/}ma", err: false},
		{v: "mama", a: 302, e: "ma{~}ma", err: false},
		{v: "mama", a: 402, e: "", err: true},
		{v: "mama", a: 105, e: "", err: true},
		{v: "mama", a: 0, e: "mama", err: false},
		{v: "ūkūs", a: 303, e: "ūkū{~}s", err: false},
	}

	for i, tc := range tests {
		v, err := ToAccentString(tc.v, tc.a)
		assert.Equal(t, tc.err, err != nil, "Fail %d", i)
		assert.Equal(t, tc.e, v, "Fail %d", i)
	}
}
