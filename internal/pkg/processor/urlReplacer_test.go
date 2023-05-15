package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReplacer(t *testing.T) {
	pr := NewURLReplacer().(*urlReplacer)
	require.NotNil(t, pr)
	assert.Equal(t, "Internetinis adresas", pr.urlPhrase)
	assert.Equal(t, "Elektroninio pašto adresas", pr.emailPhrase)
}

func TestReplacerProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.NormalizedText = []string{" a a www.delfi.lt"}
	pr := NewURLReplacer()
	require.NotNil(t, pr)
	err := pr.Process(d)
	assert.Nil(t, err)
	assert.Equal(t, " a a Internetinis adresas", d.Text)
}

func TestReplacer_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	d.CleanedText = []string{"text"}
	pr := NewURLReplacer()
	err := pr.Process(d)
	require.Nil(t, err)
	assert.Equal(t, "", d.Text)
}

func Test_replaceURLs(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "Empty", args: "", want: ""},
		{name: "No words", args: "  \n", want: "  \n"},
		{name: "Replace", args: "Olia http://delfi.lt", want: "Olia Internetinis adresas"},
		{name: "Replace", args: "Olia delfi.lt", want: "Olia Internetinis adresas"},
		{name: "Replace", args: "Olia delfi.kazkas", want: "Olia delfi.kazkas"},
		{name: "Replace several", args: "Olia www.delfi.eu\nwww.delfi.eu,:",
			want: "Olia Internetinis adresas\nInternetinis adresas,:"},
		{name: "Complex", args: "Olia https://www.delfi.eu?tatata=111&aha=tatat",
			want: "Olia Internetinis adresas"},
		{name: "Complex with point", args: "Olia https://www.delfi.eu?tatata=111&aha=tatat. Tada",
			want: "Olia Internetinis adresas. Tada"},
		{name: "Leave lrt.lt", args: "Olia lrt.lt, https://lrt.lt, www.lrt.lt/, http://www.lrt.lt/olia",
			want: "Olia lrt.lt, lrt.lt, lrt.lt, Internetinis adresas"},
		{name: "Leave vdu.lt", args: "Olia vdu.lt, https://vdu.lt/",
			want: "Olia vdu.lt, vdu.lt"},
		{name: "Leave lrs.lt", args: "Olia lrs.lt, https://lrs.lt/",
			want: "Olia lrs.lt, lrs.lt"},
		{name: "Email", args: "Olia aaa@vdu.lt, aaa.tttt.www@lrt.lt",
			want: "Olia Elektroninio pašto adresas, Elektroninio pašto adresas"},
		{name: "Email & URL", args: "Olia aaa@vdu.lt, aaa.tttt.lt",
			want: "Olia Elektroninio pašto adresas, Internetinis adresas"},
	}
	pr := NewURLReplacer().(*urlReplacer)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pr.replaceURLs(tt.args); got != tt.want {
				t.Errorf("replaceURLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_baseURL(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "drop slash", args: "www.delfi.lt/", want: "delfi.lt"},
		{name: "http", args: "http://www.delfi.lt/", want: "delfi.lt"},
		{name: "https", args: "https://www.delfi.lt/", want: "delfi.lt"},
		{name: "several", args: "https://www.delfi.lt/", want: "delfi.lt"},
		{name: "upper", args: "hTTps://wWw.Delfi.lt/", want: "Delfi.lt"},
		{name: "upper", args: "HTTP://wWw.Lrt.Lt/", want: "Lrt.Lt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := baseURL(tt.args); got != tt.want {
				t.Errorf("baseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
