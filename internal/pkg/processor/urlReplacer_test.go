package processor

import (
	"context"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	urlReaderHTTPJSONMock *mocks.HTTPInvokerJSON
)

func TestNewReplacer(t *testing.T) {
	pr := testNewURLReplacer(t)

	require.NotNil(t, pr)
}

func testNewURLReplacer(t *testing.T) *urlReplacer {
	t.Helper()
	initTestJSON(t)

	pr, err := NewURLReplacer("http://replacer.lt", "http://tagger.lt")
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
					res = append(res, TaggedWord{String: w, Mi: "Nx", Type: "WORD", Lemma: w})
				}
			}
			*params[1].(*[]TaggedWord) = res
		}).Return(nil)

	urlReaderHTTPJSONMock = &mocks.HTTPInvokerJSON{}
	pr.urlReaderHTTPWrap = urlReaderHTTPJSONMock
	urlReaderHTTPJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			input := params[0].(*urlReaderRequest)
			_ = input
			res := urlReaderResponse{}
			for _, s := range input.URLs {
				res.Items = append(res.Items, &urlChangeData{
					Text: s.Text,
					Exp: []*urlPartWord{
						{Text: "vvv", Kind: urlPartWordTypeChars},
						{Text: "taškas", Kind: urlPartWordTypeWord},
						{Text: "com", Kind: urlPartWordTypeChars},
					},
				})
			}
			*params[1].(*urlReaderResponse) = res
		}).Return(nil)
	return pr
}

func TestReplacerProcess(t *testing.T) {
	d := &synthesizer.TTSData{}
	lw := &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: miLink}}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		lw,
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "vvv", Mi: "Ys", Lemma: "vvv"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "taškas", Mi: "Nx", Lemma: "taškas"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "com", Mi: "Ys", Lemma: "com"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}, d.Words)
}

func TestReplacerProcess_email(t *testing.T) {

	lw := &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "a@elfi.lt", Mi: miEmail}}
	d := &synthesizer.TTSData{}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		lw,
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "vvv", Mi: "Ys", Lemma: "vvv"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "taškas", Mi: "Nx", Lemma: "taškas"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "com", Mi: "Ys", Lemma: "com"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
	}, d.Words)
}

func TestReplacerProcess_several(t *testing.T) {
	lw := &synthesizer.ProcessedWord{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: miLink}}
	d := &synthesizer.TTSData{}
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		lw,
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		lw,
	}
	pr := testNewURLReplacer(t)
	require.NotNil(t, pr)
	err := pr.Process(context.TODO(), d)
	require.Nil(t, err)
	assert.Equal(t, []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "vvv", Mi: "Ys", Lemma: "vvv"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "taškas", Mi: "Nx", Lemma: "taškas"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "com", Mi: "Ys", Lemma: "com"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "vvv", Mi: "Ys", Lemma: "vvv"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "taškas", Mi: "Nx", Lemma: "taškas"}, FromWord: &lw.Tagged},
		{Tagged: synthesizer.TaggedWord{Word: "com", Mi: "Ys", Lemma: "com"}, FromWord: &lw.Tagged},
	}, d.Words)
}

func TestReplacer_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	d.Words = []*synthesizer.ProcessedWord{
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: miLink}},
		{Tagged: synthesizer.TaggedWord{Word: "olia", Mi: "X-"}},
		{Tagged: synthesizer.TaggedWord{Word: "www.delfi.lt", Mi: miLink}},
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

func Test_urlFinder_replaceAll(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		text        string
		placeholder string
		want        []string
		want2       string
	}{
		{name: "No URLs", text: "Olia ooo ok", placeholder: "XXX", want: nil, want2: "Olia ooo ok"},
		{name: "One URL", text: "Olia www.delfi.lt is OK", placeholder: "XXX",
			want: []string{"www.delfi.lt"}, want2: "Olia XXX is OK"},
		{name: "Several URLs", text: "Olia www.delfi.lt is OK. Info https://lrt.lt/test?aaa=111",
			placeholder: "YYY",
			want:        []string{"www.delfi.lt", "https://lrt.lt/test?aaa=111"},
			want2:       "Olia YYY is OK. Info YYY"},
		{name: "email", text: "Olia aaa@aaa.lt and?",
			placeholder: "YYY",
			want:        []string{"aaa@aaa.lt"},
			want2:       "Olia YYY and?"},
		{name: "ip", text: "Olia 192.168.1.1 and?",
			placeholder: "YYY",
			want:        []string{"192.168.1.1"},
			want2:       "Olia YYY and?"},
	}
	p, err := NewURLFinder()
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := p.replaceAll(tt.text, tt.placeholder)
			// TODO: update the condition below to compare got with tt.want.
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceAll() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("replaceAll() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
