package ssml

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	vf := func(s string) (string, error) {
		return s, nil
	}
	tests := []struct {
		name    string
		xml     string
		vf      func(string) (string, error)
		want    []Part
		wantErr bool
	}{
		{name: "simple empty", xml: "<speak></speak>", want: []Part{}, wantErr: false},
		{name: "simple", xml: "<speak>olia</speak>", want: []Part{
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}}},
			wantErr: false},
		{name: "speak fail", xml: "aa<speak>olia</speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia</speak><speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia<speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia</speak>olia", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<b>olia</b>", want: []Part{},
			wantErr: true},
		{name: "empty", wantErr: true},
		{name: "<p>", xml: "<speak><p>olia</p></speak>", want: []Part{
			&Pause{Duration: pDuration},
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}},
			&Pause{Duration: pDuration}},
			wantErr: false},
		{name: "<p>", xml: "<speak><p>olia</p></speak>", want: []Part{
			&Pause{Duration: pDuration},
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}},
			&Pause{Duration: pDuration}},
			wantErr: false},
		{name: "<break> time", xml: `<speak><break time="10s"/></speak>`, want: []Part{
			&Pause{Duration: time.Second * 10, IsBreak: true}},
			wantErr: false},
		{name: "<break> strength", xml: `<speak><break strength="x-weak"/></speak>`, want: []Part{
			&Pause{Duration: time.Millisecond * 250, IsBreak: true},
		}, wantErr: false},
		{name: "<break> with text", xml: `<speak><break strength="x-weak">olia</break></speak>`,
			want: []Part{}, wantErr: true},
		{name: "<voice> strength", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<voice> inner", xml: `<speak>
		<voice name="ooo">aaa
		<voice name="ooo1">aaa1</voice>
		end
		</voice>
		end def</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 1, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<prosody> rate", xml: `<speak><prosody rate="50%">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 2, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> inner", xml: `<speak><prosody rate="200%">
		<voice name="ooo">aaa
		<voice name="ooo1"><prosody rate="slow">aaa1</prosody></voice>
		end
		</voice></prosody>
		end def</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 0.5, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 1.5, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 0.5, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<voice> map", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf: func(s string) (string, error) { return "ooo.v1", nil },
			want: []Part{
				&Text{Voice: "ooo.v1", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<voice> map fail", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf:   func(s string) (string, error) { return "", errors.New("no voice") },
			want: []Part{}, wantErr: true},
		{name: "syntax error", xml: "<speak>olia</sss>", want: []Part{},
			wantErr: true},
		{name: "speak empty text", xml: "<speak></speak>   ", want: []Part{},
			wantErr: false},
		{name: "speak empty text", xml: "<speak></speak> \n\n\n  ", want: []Part{},
			wantErr: false},
		{name: "<w> ", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<w> joins", xml: `<speak>olia<intelektika:w acc="g{a/}li">gali</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}, {Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<w> several", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w>
		<intelektika:w acc="{a/}li">gali</intelektika:w>
		</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}, {Text: "gali", Accented: "{a/}li"}}},
			}, wantErr: false},
		{name: "<w> splits", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w><p/><intelektika:w acc="g{a/}li">gali</intelektika:w>
			</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
				&Pause{Duration: pDuration},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<voice> splits", xml: `<speak>text1<voice name="v1">gali</voice>text2</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text1"}}},
				&Text{Voice: "v1", Speed: 1, Texts: []TextPart{{Text: "gali"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text2"}}},
			}, wantErr: false},
		{name: "joins comment", xml: `<speak>text1
			<!-- comment -->
			text2</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text1"}, {Text: "text2"}}},
			}, wantErr: false},
		{name: "fail <w> in <w>", xml: `<speak><intelektika:w acc="g{a/}li"><intelektika:w acc="g{a/}li">gali</intelektika:w></intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> no text", xml: `<speak><intelektika:w acc="g{a/}li"></intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> no acc", xml: `<speak><intelektika:w acc="">gali</intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> in <break>", xml: `<speak><break time="10s"><intelektika:w acc="g{a/}li">gali</intelektika:w></break></speak>`,
			want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &Text{Voice: "aa", Speed: 1}
			vf := vf
			if tt.vf != nil {
				vf = tt.vf
			}
			got, err := Parse(strings.NewReader(tt.xml), def, vf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("Parse() = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("Parse() = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func Test_getDuration(t *testing.T) {
	type args struct {
		tm  string
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{name: "time", args: args{tm: "10s", str: ""}, want: time.Second * 10, wantErr: false},
		{name: "time", args: args{tm: "10ms", str: ""}, want: time.Millisecond * 10, wantErr: false},
		{name: "time", args: args{tm: "1s250ms", str: ""}, want: time.Millisecond * 1250, wantErr: false},
		{name: "strength", args: args{tm: "", str: "none"}, want: time.Millisecond * 0, wantErr: false},
		{name: "strength", args: args{tm: "", str: "x-weak"}, want: time.Millisecond * 250, wantErr: false},
		{name: "strength", args: args{tm: "", str: "weak"}, want: time.Millisecond * 500, wantErr: false},
		{name: "strength", args: args{tm: "", str: "medium"}, want: time.Millisecond * 750, wantErr: false},
		{name: "strength", args: args{tm: "", str: "strong"}, want: time.Millisecond * 1000, wantErr: false},
		{name: "strength", args: args{tm: "", str: "x-strong"}, want: time.Millisecond * 1250, wantErr: false},
		{name: "strength fail", args: args{tm: "", str: "aaa"}, want: time.Millisecond * 0, wantErr: true},
		{name: "none", args: args{tm: "", str: ""}, want: time.Millisecond * 0, wantErr: true},
		{name: "time fail", args: args{tm: "xxx", str: ""}, want: time.Millisecond * 0, wantErr: true},
		{name: "both fail", args: args{tm: "xxx", str: "aaa"}, want: time.Millisecond * 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDuration(tt.args.tm, tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSpeed(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    float32
		wantErr bool
	}{
		{name: "percent", args: "10%", want: 2, wantErr: false},
		{name: "percent", args: "-40%", want: 2, wantErr: false},
		{name: "percent", args: "50%", want: 2, wantErr: false},
		{name: "percent", args: "75%", want: 1.5, wantErr: false},
		{name: "percent", args: "100%", want: 1, wantErr: false},
		{name: "percent", args: "120%", want: .9, wantErr: false},
		{name: "percent", args: "150%", want: .75, wantErr: false},
		{name: "percent", args: "200%", want: .5, wantErr: false},
		{name: "percent", args: "300%", want: .5, wantErr: false},
		{name: "percent fail", args: "300", want: 0, wantErr: true},
		{name: "percent fail", args: "aa%", want: 0, wantErr: true},
		{name: "rate", args: "x-slow", want: 2, wantErr: false},
		{name: "rate", args: "slow", want: 1.5, wantErr: false},
		{name: "rate", args: "medium", want: 1, wantErr: false},
		{name: "rate", args: "default", want: 1, wantErr: false},
		{name: "rate", args: "fast", want: .75, wantErr: false},
		{name: "rate", args: "x-fast", want: .5, wantErr: false},
		{name: "rate fail", args: "aaa x-slow", want: 0, wantErr: true},
		{name: "empty fail", args: "", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSpeed(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSpeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_okAccentedWord(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{name: "word", args: "aaa", want: true},
		{name: "acc", args: "{a/}", want: true},
		{name: "empty", args: " ", want: false},
		{name: "empty", args: "", want: false},
		{name: "long", args: strings.Repeat("a", 50), want: false},
		{name: "wrong acc", args: "{a-}", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := okAccentedWord(tt.args); got != tt.want {
				t.Errorf("okAccentedWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
