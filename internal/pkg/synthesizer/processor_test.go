package synthesizer

import (
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
	"github.com/airenas/tts-line/pkg/ssml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestWork_SSML(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.TextWithNumbers = "olia lia"
		assert.Equal(t, 3, len(d.SSMLParts))
		return nil
	}
	worker.processors = nil
	worker.AddSSML(processorMock)
	res, err := worker.Work(&api.TTSRequestConfig{Text: "<speak>olia</speak>", OutputTextFormat: api.TextNormalized,
		SSMLParts: []ssml.Part{&ssml.Text{Texts: []ssml.TextPart{{Text: "Olia"}}}, &ssml.Text{Texts: []ssml.TextPart{{Text: "Olia"}}},
			&ssml.Pause{Duration: 10 * time.Second}}})
	assert.Nil(t, err)
	assert.Equal(t, "olia lia", res.Text)
}

func TestWork_SSML_Fail(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		return errors.New("fail")
	}
	worker.processors = nil
	worker.AddSSML(processorMock)
	_, err := worker.Work(&api.TTSRequestConfig{Text: "<speak>olia</speak>", OutputTextFormat: api.TextNormalized,
		SSMLParts: []ssml.Part{&ssml.Text{Texts: []ssml.TextPart{{Text: "Olia"}}}}})
	assert.NotNil(t, err)
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

func TestGetTranscriberAccent(t *testing.T) {
	assert.Equal(t, 101, GetTranscriberAccent(&ProcessedWord{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}}))
	assert.Equal(t, 103, GetTranscriberAccent(&ProcessedWord{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}, UserAccent: 103}))
	assert.Equal(t, 101, GetTranscriberAccent(&ProcessedWord{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}, Clitic: Clitic{Type: CliticsUnused}}))
	assert.Equal(t, 0, GetTranscriberAccent(&ProcessedWord{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}, Clitic: Clitic{Type: CliticsNone}}))
	assert.Equal(t, 105, GetTranscriberAccent(&ProcessedWord{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 101}, Clitic: Clitic{Type: CliticsCustom, Accent: 105}}))
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

func Test_makeSSMLParts(t *testing.T) {
	tests := []struct {
		name    string
		args    *api.TTSRequestConfig
		want    []*TTSData
		wantErr bool
	}{
		{name: "text", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Text{Voice: "aa", Texts: []ssml.TextPart{{Text: "oo"}}, Speed: 0.6}}},
			want:    []*TTSData{{OriginalTextParts: []*TTSTextPart{{Text: "oo"}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa", Speed: 0.6}}},
			wantErr: false},
		{name: "pause", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Pause{Duration: time.Second, IsBreak: true}}},
			want:    []*TTSData{{Cfg: TTSConfig{Type: SSMLPause, PauseDuration: time.Second}}},
			wantErr: false},
		{name: "pause", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Pause{Duration: time.Second, IsBreak: true},
			&ssml.Text{Voice: "aa", Texts: []ssml.TextPart{{Text: "oo"}}, Speed: 0.6}}},
			want: []*TTSData{{Cfg: TTSConfig{Type: SSMLPause, PauseDuration: time.Second}},
				{OriginalTextParts: []*TTSTextPart{{Text: "oo"}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa", Speed: 0.6}}},
			wantErr: false},
		{name: "fail", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{struct{}{}}},
			want:    []*TTSData{},
			wantErr: true},
		{name: "text with accent", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Text{Voice: "aa", Texts: []ssml.TextPart{{Text: "oo", Accented: "acc"}}, Speed: 0.6}}},
			want:    []*TTSData{{OriginalTextParts: []*TTSTextPart{{Text: "oo", Accented: "acc"}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa", Speed: 0.6}}},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeSSMLParts(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSSMLParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, len(tt.want), len(got))
			for i := range tt.want {
				assert.Equal(t, tt.want[i].OriginalTextParts, got[i].OriginalTextParts)
				assert.Equal(t, tt.want[i].Cfg.Voice, got[i].Cfg.Voice)
				assert.Equal(t, tt.want[i].Cfg.Speed, got[i].Cfg.Speed)
				assert.Equal(t, tt.want[i].Cfg.Type, got[i].Cfg.Type)
			}
		})
	}
}
