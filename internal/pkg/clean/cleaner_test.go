package clean

import (
	"testing"

	"gotest.tools/assert"
)

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		v string
		e string
		i string
	}{
		{v: "olia<p>tata<b></p>", e: "olia\ntata"},
		{v: "", e: ""},
		{v: "<b>olia<l>tata<b>", e: "olia tata"},
		{v: "<em>olia<i>", e: "olia"},
		{v: "mama    ir jis\n\no, kas", e: "mama ir jis\no, kas"},
		{v: "mama <strict>ir</strict><b> jis</b><i>no no</i>", e: "mama ir jis no no"},
		{v: "olia<a = href\"https://semantika.lt\">kur</a>", e: "olia kur"},
		{v: "<div><p>tada\n\nkur\n no\n</p></div>", e: "tada\nkur\nno"},
		{v: "olia http://olia.lt", e: "olia http://olia.lt", i: "Leaves link"},
		{v: "&lt; &gt; &amp;", e: "< > &", i: "Changes symbols &lt &gt &"},
	}

	for i, tc := range tests {
		assert.Equal(t, tc.e, cleanHTML(tc.v), "Fail %d - %s", i, tc.i)
	}
}
