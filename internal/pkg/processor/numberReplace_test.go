package processor

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	httpInvokerMock *mockHTTPInvoker
)

func initTest(t *testing.T) {
	httpInvokerMock = &mockHTTPInvoker{}
}

func TestCreateNumberReplace(t *testing.T) {
	initTest(t)
	pr, err := NewNumberReplace("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateNumberReplace_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewNumberReplace("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeNumberReplace(t *testing.T) {
	initTest(t)
	pr, _ := NewNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*numberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{Text: []string{"3"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "trys"
		}).Return(nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"trys"}, d.TextWithNumbers)
}

func TestInvokeNumberReplace_Fail(t *testing.T) {
	initTest(t)
	pr, _ := NewNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*numberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(&d)
	assert.NotNil(t, err)
}

func TestInvokeNumberReplace_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewNumberReplace("http://server")
	err := pr.Process(d)
	assert.Nil(t, err)
}

func TestCreateSSMLNumberReplace(t *testing.T) {
	initTest(t)
	pr, err := NewSSMLNumberReplace("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateSSMLNumberReplace_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewSSMLNumberReplace("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeSSMLNumberReplace(t *testing.T) {
	initTest(t)
	pr, _ := NewSSMLNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{Text: []string{"3 oli{a/} 4"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "trys olia keturi"
		}).Return(nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"trys oli{a/} keturi"}, d.TextWithNumbers)
}

func TestInvokeSSMLNumberReplace_Real(t *testing.T) {
	initTest(t)
	pr, _ := NewSSMLNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{Text: []string{"kadenciją 1998-2002 m. {o\\}rb buvo demokratiškas vadovas. Kad 2010 m. grįžęs į valdžią jis staiga virto autokratu, buvo didelis"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais orb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo didelis"
		}).Return(nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais {o\\}rb buvo demokratiškas vadovas." +
		" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo didelis"}, d.TextWithNumbers)
}

func TestInvokeSSMLNumberReplace_Real1(t *testing.T) {
	initTest(t)
	pr, _ := NewSSMLNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{Text: []string{"kadenciją 1998-2002 m. {o\\}rb buvo demokratiškas vadovas. Kad 2010 m. grįžęs į valdžią jis staiga virto autokratu, buvo"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais orb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo"
		}).Return(nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais {o\\}rb buvo demokratiškas vadovas." +
		" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo"}, d.TextWithNumbers)
}

func TestInvokeSSMLNumberReplace_Real2(t *testing.T) {
	initTest(t)
	pr, _ := NewSSMLNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{Text: []string{"Robertsonas (1988 m. surinkęs 25 proc. balsų), bjūk{e~}nenas (1996 m. surinkęs 23 proc. balsų) ir f{o/}rbzas (2000 m. surinkęs 31 proc. balsų)"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "Robertsonas (tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūkenenas (tūkstantis devyni šimtai " +
				"devyniasdešimt šeštieji metai surinkęs dvidešimt trys procentai balsų) ir forbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)"
		}).Return(nil)
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"Robertsonas (tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūk{e~}nenas (tūkstantis devyni šimtai devyniasdešimt " +
		"šeštieji metai surinkęs dvidešimt trys procentai balsų) ir f{o/}rbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)"}, d.TextWithNumbers)
}

func Test_doPartlyAlign(t *testing.T) {
	type args struct {
		s1 []string
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{name: "map last", args: args{s1: []string{"a", "6", "c", "d"}, s2: []string{"a", "b", "c", "c", "d"}}, want: []int{0, 1, 3, 4}},
		{name: "insert", args: args{s1: []string{"a", "6", "c"}, s2: []string{"a", "b", "b1", "c"}}, want: []int{0, 1, 3}},
		{name: "same", args: args{s1: []string{"a", "b", "c"}, s2: []string{"a", "b", "c"}}, want: []int{0, 1, 2}},
		{name: "change", args: args{s1: []string{"a", "6", "c"}, s2: []string{"a", "b", "c"}}, want: []int{0, 1, 2}},
		{name: "insert several", args: args{s1: []string{"a", "6", "c"}, s2: []string{"a", "b", "b1", "b1", "c"}}, want: []int{0, 1, 4}},
		{name: "insert several", args: args{s1: []string{"a", "6", "c", "7"}, s2: []string{"a", "b", "b1", "b1", "c", "d", "d"}},
			want: []int{0, 1, 4, 5}},
		{name: "skip", args: args{s1: []string{"a", "6", "c"}, s2: []string{"a", "c"}}, want: []int{0, -1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doPartlyAlign(tt.args.s1, tt.args.s2, 20); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("doPartlyAlign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapAccentsBack(t *testing.T) {
	type args struct {
		new  string
		orig []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "empty", args: args{new: "", orig: []string{""}}, want: []string{""}, wantErr: false},
		{name: "no accent", args: args{new: "a b c d r", orig: []string{"a b c d r"}}, want: []string{"a b c d r"}, wantErr: false},
		{name: "with accent", args: args{new: "a b c d r", orig: []string{"a {b/} c d {r~}"}}, want: []string{"a {b/} c d {r~}"}, wantErr: false},
		{name: "fail on other word", args: args{new: "a b c d k", orig: []string{"a {b/} c d {r~}"}}, want: nil, wantErr: true},
		{name: "fail on missing", args: args{new: "a b c d", orig: []string{"a {b/} c d {r~}"}}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapAccentsBack(tt.args.new, tt.args.orig)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapAccentsBack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapAccentsBack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findLastConsequtiveAlign(t *testing.T) {
	type args struct {
		partlyAll []int
		oStrs     []string
		nStrs     []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "ok - the last", args: args{partlyAll: []int{0, 1, 2, 3, 4, 5}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "d", "e", "f"}}, want: 5, wantErr: false},
		{name: "ok - the last shifted", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "i", "d", "e", "f"}}, want: 5, wantErr: false},
		{name: "ok - not the last", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "i", "d", "e", "fc"}}, want: 4, wantErr: false},
		{name: "fails match", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b,", "c,", "i", "d", "e,", "f,"}}, want: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findLastConsequtiveAlign(tt.args.partlyAll, tt.args.oStrs, tt.args.nStrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("findLastConsequtiveAlign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findLastConsequtiveAlign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_align(t *testing.T) {
	type args struct {
		oStrs []string
		nStrs []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{name: "with shift", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: append([]int{0}, makeTestInt(6, 105)...), wantErr: false},
		{name: "simple", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "c", "d"}},
			want: []int{0, 1, 2, 3}, wantErr: false},
		{name: "insert", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "b1", "c", "d"}},
			want: []int{0, 1, 3, 4}, wantErr: false},
		{name: "skip", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "c", "d"}},
			want: []int{0, -1, 1, 2}, wantErr: false},
		{name: "long", args: args{oStrs: makeTestStr("a", 1000), nStrs: makeTestStr("a", 1000)},
			want: makeTestInt(0, 1000), wantErr: false},
		{name: "super long", args: args{oStrs: makeTestStr("a", 10000), nStrs: makeTestStr("a", 10000)},
			want: makeTestInt(0, 10000), wantErr: false},
		{name: "fails", args: args{oStrs: makeTestStr("a", 30), nStrs: makeTestStr("b", 30)},
			want: nil, wantErr: true},
		{name: "with shift", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: append([]int{0}, makeTestInt(6, 105)...), wantErr: false},
		{name: "with in middle", args: args{
			oStrs: append(makeTestStr("a", 100), makeTestStr("a", 100)...),
			nStrs: append(append(makeTestStr("a", 100), makeTestStr("c", 5)...), makeTestStr("a", 100)...)},
			want: append(makeTestInt(0, 100), makeTestInt(105, 205)...), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := align(tt.args.oStrs, tt.args.nStrs, 20)
			if (err != nil) != tt.wantErr {
				t.Errorf("align() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("align() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeTestStr(pr string, n int) []string {
	res := []string{}
	for i := 0; i < n; i++ {
		res = append(res, fmt.Sprintf("%s-%d", pr, i))
	}
	return res
}

func makeTestInt(from, n int) []int {
	res := []int{}
	for i := from; i < n; i++ {
		res = append(res, i)
	}
	return res
}

type mockHTTPInvoker struct{ mock.Mock }

func (m *mockHTTPInvoker) InvokeText(in string, out interface{}) error {
	args := m.Called(in, out)
	return args.Error(0)
}
