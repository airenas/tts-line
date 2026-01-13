package processor

import (
	"context"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewReplacer(t *testing.T) {
	pr := testNewURLReplacer(t)
	assert.Equal(t, "Internetinis adresas", pr.urlPhrase)
	assert.Equal(t, "Elektroninio pašto adresas", pr.emailPhrase)
}

func testNewURLReplacer(t *testing.T) *urlReplacer {
	t.Helper()
	initTestJSON(t)

	pr, err := NewURLReplacer("http://tagger.lt")
	require.Nil(t, err)
	require.NotNil(t, pr)

	// mimic tagger to return all words as normal words
	pr.taggerHTTPWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			input := params[0].([][]string)
			res := []TaggedWord{}
			for _, s := range input {
				for _, w := range s {
					res = append(res, TaggedWord{String: w, Mi: "Nx", Type: "WORD"})
				}
			}
			*params[1].(*[]TaggedWord) = res
		}).Return(nil)
	return pr
}

func TestReplacerProcess(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: linkMI}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "Internetinis", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "adresas", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}, d.Words)
}

func TestReplacerProcess_email(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "a@delfi.lt", Mi: emailMI}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "Elektroninio", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "pašto", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "adresas", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}, d.Words)
}

func TestReplacerProcess_several(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: linkMI}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: linkMI}},
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "Internetinis", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "adresas", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "Internetinis", Mi: "Nx"}},
		{Tagged: synthesizer.TaggedWord{Word: "adresas", Mi: "Nx"}},
	}, d.Words)
}

func TestReplacer_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: linkMI}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: linkMI}},
	}
	pr := testNewURLReplacer(t)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, 4, len(d.Words))
}

// func Test_replaceURLs(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		args string
// 		want string
// 	}{
// 		{name: "Empty", args: "", want: ""},
// 		{name: "No words", args: "  \n", want: "  \n"},
// 		{name: "Replace", args: "Olia http://delfi.lt", want: "Olia Internetinis adresas"},
// 		{name: "Replace", args: "Olia delfi.lt", want: "Olia Internetinis adresas"},
// 		{name: "Replace", args: "Olia delfi.kazkas", want: "Olia delfi.kazkas"},
// 		{name: "Replace several", args: "Olia www.delfi.eu\nwww.delfi.eu,:",
// 			want: "Olia Internetinis adresas\nInternetinis adresas,:"},
// 		{name: "Complex", args: "Olia https://www.delfi.eu?tatata=111&aha=tatat",
// 			want: "Olia Internetinis adresas"},
// 		{name: "Complex with point", args: "Olia https://www.delfi.eu?tatata=111&aha=tatat. Tada",
// 			want: "Olia Internetinis adresas. Tada"},
// 		{name: "Leave lrt.lt", args: "Olia lrt.lt, https://lrt.lt, www.lrt.lt/, http://www.lrt.lt/olia",
// 			want: "Olia lrt.lt, lrt.lt, lrt.lt, Internetinis adresas"},
// 		{name: "Leave vdu.lt", args: "Olia vdu.lt, https://vdu.lt/",
// 			want: "Olia vdu.lt, vdu.lt"},
// 		{name: "Leave lrs.lt", args: "Olia lrs.lt, https://lrs.lt/",
// 			want: "Olia lrs.lt, lrs.lt"},
// 		{name: "Email", args: "Olia aaa@vdu.lt, aaa.tttt.www@lrt.lt",
// 			want: "Olia Elektroninio pašto adresas, Elektroninio pašto adresas"},
// 		{name: "Email & URL", args: "Olia aaa@vdu.lt, aaa.tttt.lt",
// 			want: "Olia Elektroninio pašto adresas, Internetinis adresas"},
// 	}
// 	pr := testNewURLReplacer(t)
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := pr.replaceURLs(t.Context(), tt.args)
// 			if got := pr.replaceURLs(t.Context(), tt.args); got != tt.want {
// 				t.Errorf("replaceURLs() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

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
