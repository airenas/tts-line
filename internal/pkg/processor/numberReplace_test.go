package processor

import (
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	httpInvokerMock *mocks.MockHTTPInvoker
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	httpInvokerMock = mocks.NewMockHTTPInvoker()
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
	d := synthesizer.TTSData{}
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			*params[1].(*string) = "trys"
			return []pegomock.ReturnValue{nil}
		})
	err := pr.Process(&d)
	assert.Nil(t, err)
	assert.Equal(t, "trys", d.TextWithNumbers)
}

func TestInvokeNumberReplace_Fail(t *testing.T) {
	initTest(t)
	pr, _ := NewNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*numberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	pegomock.When(httpInvokerMock.InvokeText(pegomock.AnyString(), pegomock.AnyInterface())).ThenReturn(errors.New("haha"))
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

func Test_doPartlyAllign(t *testing.T) {
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
			if got := doPartlyAllign(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("doPartlyAllign() = %v, want %v", got, tt.want)
			}
		})
	}
}
