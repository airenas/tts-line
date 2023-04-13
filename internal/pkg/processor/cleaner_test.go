package processor

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCleaner(t *testing.T) {
	initTestJSON(t)
	pr, err := NewCleaner("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewCleaner_Fails(t *testing.T) {
	initTestJSON(t)
	pr, err := NewCleaner("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestCleanProcess(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*normData) = normData{Text: "clean text"}
		}).Return(nil)
	err := pr.Process(d)
	assert.Nil(t, err)
	httpJSONMock.AssertNumberOfCalls(t, "InvokeJSON", 1)
	cp1 := httpJSONMock.Calls[0].Arguments[0]
	assert.Equal(t, &normData{Text: " a a"}, cp1)
	assert.Equal(t, "clean text", d.CleanedText)
}

func TestCleanProcess_Fail(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Return(errors.New("olia"))
	err := pr.Process(d)
	assert.NotNil(t, err)
}

func TestCleanProcess_NoText(t *testing.T) {
	initTestJSON(t)
	d := &synthesizer.TTSData{}
	d.OriginalText = " a a"
	pr, _ := NewCleaner("http://server")
	assert.NotNil(t, pr)
	pr.(*cleaner).httpWrap = httpJSONMock
	httpJSONMock.On("InvokeJSON", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*normData) = normData{Text: ""}
		}).Return(nil)
	err := pr.Process(d)
	assert.Equal(t, utils.ErrNoInput, err)
}

func TestClean_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewCleaner("http://server")
	err := pr.Process(d)
	assert.Nil(t, err)
}

func Test_getNormText(t *testing.T) {
	tests := []struct {
		name string
		args *synthesizer.TTSData
		want string
	}{
		{name: "originalText", args: &synthesizer.TTSData{OriginalText: "olia olia"}, want: "olia olia"},
		{name: "from Parts", args: &synthesizer.TTSData{OriginalTextParts: []*synthesizer.TTSTextPart{{Text: "olia"}, {Text: "olia"}}}, want: "olia olia"},
		{name: "from Parts Accented", args: &synthesizer.TTSData{OriginalTextParts: []*synthesizer.TTSTextPart{{Text: "olia"}, {Text: "o1", Accented: "aa{a/}"}}}, want: "olia aa{a/}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNormText(tt.args); got != tt.want {
				t.Errorf("getNormText() = %v, want %v", got, tt.want)
			}
		})
	}
}
