package synthesizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

var (
	processorMock *procMock
	worker        *MainWorker
)

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	processorMock = &procMock{f: func(d *TTSData) error { return nil }}
	worker = &MainWorker{}
	worker.Add(processorMock)
}

func TestWork(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		assert.Equal(t, "olia", d.OriginalText)
		d.AudioMP3 = "mp3"
		return nil
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "mp3", res.AudioAsString)
}

func TestWork_Fails(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		return errors.New("olia")
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWork_ValidationFailure(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.ValidationFailures = append(d.ValidationFailures, api.ValidateFailure{FailingPosition: 10})
		return nil
	}
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "", res.AudioAsString)
	assert.Equal(t, 10, res.ValidationFailures[0].FailingPosition)
}

func TestWork_Several(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.AudioMP3 = "wav"
		return nil
	}
	processorMock1 := &procMock{f: func(d *TTSData) error {
		d.AudioMP3 = d.AudioMP3 + "mp3"
		return nil
	}}
	worker.Add(processorMock1)
	res, _ := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Equal(t, "wavmp3", res.AudioAsString)
}

func TestWork_StopProcess(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.ValidationFailures = append(d.ValidationFailures, api.ValidateFailure{FailingPosition: 10})
		return nil
	}
	processorMock1 := &procMock{f: func(d *TTSData) error {
		assert.Fail(t, "Unexpected call")
		return nil
	}}
	worker.Add(processorMock1)
	res, err := worker.Work(&api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestWork_HasUUID(t *testing.T) {
	initTest(t)
	res, _ := worker.Work(&api.TTSRequestConfig{Text: "olia", AllowCollectData: true, OutputTextFormat: api.TextNormalized})
	assert.NotEqual(t, "", res.RequestID)
	res, _ = worker.Work(&api.TTSRequestConfig{Text: "olia", AllowCollectData: true, OutputTextFormat: api.TextNone})
	assert.Equal(t, "", res.RequestID)
	res, _ = worker.Work(&api.TTSRequestConfig{Text: "olia", AllowCollectData: false, OutputTextFormat: api.TextNormalized})
	assert.Equal(t, "", res.RequestID)
}

func TestWork_ReturnText(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.TextWithNumbers = "olia lia"
		return nil
	}
	res, _ := worker.Work(&api.TTSRequestConfig{Text: "olia", OutputTextFormat: api.TextNormalized})
	assert.Equal(t, "olia lia", res.Text)
	res, _ = worker.Work(&api.TTSRequestConfig{Text: "olia", OutputTextFormat: api.TextNone})
	assert.Equal(t, "", res.Text)
}

func TestMapResult_Accented(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextAccented}
	d.Parts = []*TTSDataPart{{Words: []*ProcessedWord{{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}},
		{Tagged: TaggedWord{Space: true}}, {Tagged: TaggedWord{Separator: ","}},
		{Tagged: TaggedWord{Word: "ai"}, AccentVariant: &AccentVariant{Accent: 302}}}}}
	res, err := mapResult(d)
	assert.Nil(t, err)
	assert.Equal(t, "{a\\}a ,a{i~}", res.Text)
}

func TestMapResult_Normalized(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextNormalized}
	d.TextWithNumbers = "oo"
	res, err := mapResult(d)
	assert.Nil(t, err)
	assert.Equal(t, "oo", res.Text)
}

func TestMapResult_AccentedFail(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextAccented}
	d.Parts = []*TTSDataPart{{Words: []*ProcessedWord{{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 401}}}}}
	_, err := mapResult(d)
	assert.NotNil(t, err)
}

func TestMapResult_FailOutputTextType(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextFormatEnum(10)}
	_, err := mapResult(d)
	assert.NotNil(t, err)
}

type procMock struct {
	f func(res *TTSData) error
}

func (pr *procMock) Process(d *TTSData) error {
	return pr.f(d)
}
