package synthesizer

import (
	"context"
	"testing"
	"time"

	"github.com/airenas/tts-line/internal/pkg/service/api"
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
	processorMock = &procMock{f: func(d *TTSData) error { return nil }}
	worker = &MainWorker{}
	worker.Add(processorMock)
}

func TestWork(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		assert.Equal(t, "olia", d.OriginalText)
		d.AudioMP3 = []byte("mp3")
		return nil
	}
	res, err := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "mp3", string(res.Audio))
}

func TestWork_Fails(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		return errors.New("olia")
	}
	res, err := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestWork_Several(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.AudioMP3 = []byte("wav")
		return nil
	}
	processorMock1 := &procMock{f: func(d *TTSData) error {
		d.AudioMP3 = append(d.AudioMP3, []byte("mp3")...)
		return nil
	}}
	worker.Add(processorMock1)
	res, _ := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia"})
	assert.Equal(t, "wavmp3", string(res.Audio))
}

func TestWork_HasUUID(t *testing.T) {
	initTest(t)
	res, _ := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia", AllowCollectData: true, OutputTextFormat: api.TextNormalized})
	assert.NotEqual(t, "", res.RequestID)
	res, _ = worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia", AllowCollectData: true, OutputTextFormat: api.TextNone})
	assert.Equal(t, "", res.RequestID)
	res, _ = worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia", AllowCollectData: false, OutputTextFormat: api.TextNormalized})
	assert.Equal(t, "", res.RequestID)
}

func TestWork_ReturnText(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.TextWithNumbers = []string{"olia lia"}
		return nil
	}
	res, _ := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia", OutputTextFormat: api.TextNormalized})
	assert.Equal(t, "olia lia", res.Text)
	res, _ = worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "olia", OutputTextFormat: api.TextNone})
	assert.Equal(t, "", res.Text)
}

func TestWork_SSML(t *testing.T) {
	initTest(t)
	processorMock.f = func(d *TTSData) error {
		d.TextWithNumbers = []string{"olia lia"}
		assert.Equal(t, 3, len(d.SSMLParts))
		return nil
	}
	worker.processors = nil
	worker.AddSSML(processorMock)
	res, err := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "<speak>olia</speak>", OutputTextFormat: api.TextNormalized,
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
	_, err := worker.Work(context.TODO(), &api.TTSRequestConfig{Text: "<speak>olia</speak>", OutputTextFormat: api.TextNormalized,
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
	res, err := mapResult(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "{a\\}a ,a{i~}", res.Text)
}

func TestMapResult_Normalized(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextNormalized}
	d.TextWithNumbers = []string{"oo"}
	res, err := mapResult(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "oo", res.Text)
}

func TestMapResult_Transcribed(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextTranscribed}
	d.Parts = []*TTSDataPart{{TranscribedText: "a b c sil"}, {TranscribedText: "d sil"}}
	res, err := mapResult(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "a b c sil d sil", res.Text)
}

func TestMapResult_Transcribed_SSML(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextTranscribed}
	d.SSMLParts = []*TTSData{{Parts: []*TTSDataPart{{TranscribedText: "a b c sil"}, {TranscribedText: "d sil"}}},
		{Parts: []*TTSDataPart{{TranscribedText: "c , sil"}}}}
	res, err := mapResult(context.TODO(), d)
	assert.Nil(t, err)
	assert.Equal(t, "a b c sil d sil c , sil", res.Text)
}

func TestMapResult_AccentedFail(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextAccented}
	d.Parts = []*TTSDataPart{{Words: []*ProcessedWord{{Tagged: TaggedWord{Word: "aa"},
		AccentVariant: &AccentVariant{Accent: 401}}}}}
	_, err := mapResult(context.TODO(), d)
	assert.NotNil(t, err)
}

func TestMapResult_FailOutputTextType(t *testing.T) {
	d := &TTSData{}
	d.Input = &api.TTSRequestConfig{OutputTextFormat: api.TextFormatEnum(10)}
	_, err := mapResult(context.TODO(), d)
	assert.NotNil(t, err)
}

type procMock struct {
	f func(res *TTSData) error
}

func (pr *procMock) Process(ctx context.Context, d *TTSData) error {
	return pr.f(d)
}

func Test_makeSSMLParts(t *testing.T) {
	tests := []struct {
		name    string
		args    *api.TTSRequestConfig
		want    []*TTSData
		wantErr bool
	}{
		{name: "text", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Text{Voice: "aa", Texts: []ssml.TextPart{{Text: "oo"}}, Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}},
			want:    []*TTSData{{OriginalTextParts: []*TTSTextPart{{Text: "oo", Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa"}}},
			wantErr: false},
		{name: "pause", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Pause{Duration: time.Second, IsBreak: true}}},
			want:    []*TTSData{{Cfg: TTSConfig{Type: SSMLPause, PauseDuration: time.Second}}},
			wantErr: false},
		{name: "pause", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Pause{Duration: time.Second, IsBreak: true},
			&ssml.Text{Voice: "aa", Texts: []ssml.TextPart{{Text: "oo"}}, Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}},
			want: []*TTSData{{Cfg: TTSConfig{Type: SSMLPause, PauseDuration: time.Second}},
				{OriginalTextParts: []*TTSTextPart{{Text: "oo", Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa"}}},
			wantErr: false},
		{name: "fail", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{struct{}{}}},
			want:    []*TTSData{},
			wantErr: true},
		{name: "text with accent", args: &api.TTSRequestConfig{SSMLParts: []ssml.Part{&ssml.Text{Voice: "aa",
			Texts: []ssml.TextPart{{Text: "oo", Accented: "acc"}}, Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}},
			want:    []*TTSData{{OriginalTextParts: []*TTSTextPart{{Text: "oo", Accented: "acc", Prosodies: []*ssml.Prosody{{Rate: 0.6}}}}, Cfg: TTSConfig{Type: SSMLText, Voice: "aa"}}},
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
				// assert.Equal(t, tt.want[i].Cfg.Prosodies, got[i].Cfg.Prosodies)
				assert.Equal(t, tt.want[i].Cfg.Type, got[i].Cfg.Type)
			}
		})
	}
}

func TestGetTranscriberAccent(t *testing.T) {
	type args struct {
		w *ProcessedWord
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "from accenter", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"}, AccentVariant: &AccentVariant{Accent: 101}}}, want: 101},
		{name: "from user", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"}, AccentVariant: &AccentVariant{Accent: 101}, UserAccent: 103}}, want: 103},
		{name: "clitics unused", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"}, AccentVariant: &AccentVariant{Accent: 101},
			Clitic: Clitic{Type: CliticsUnused}}}, want: 101},
		{name: "clitics none", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"}, AccentVariant: &AccentVariant{Accent: 101},
			Clitic: Clitic{Type: CliticsNone}}}, want: 0},
		{name: "clitics custom", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"}, AccentVariant: &AccentVariant{Accent: 101},
			Clitic: Clitic{Type: CliticsCustom, Accent: 105}}}, want: 105},
		{name: "no accent from user", args: args{w: &ProcessedWord{Tagged: TaggedWord{Word: "aa"},
			AccentVariant: &AccentVariant{Accent: 101}, TextPart: &TTSTextPart{Accented: "aa"}}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTranscriberAccent(tt.args.w); got != tt.want {
				t.Errorf("GetTranscriberAccent() = %v, want %v", got, tt.want)
			}
		})
	}
}
