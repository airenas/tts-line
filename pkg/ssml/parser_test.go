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
			&Text{Text: "olia", Voice: "aa", Speed: 1}},
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
			&Text{Text: "olia", Voice: "aa", Speed: 1},
			&Pause{Duration: pDuration}},
			wantErr: false},
		{name: "<p>", xml: "<speak><p>olia</p></speak>", want: []Part{
			&Pause{Duration: pDuration},
			&Text{Text: "olia", Voice: "aa", Speed: 1},
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
				&Text{Text: "aaa", Voice: "ooo", Speed: 1},
			}, wantErr: false},
		{name: "<voice> inner", xml: `<speak>
		<voice name="ooo">aaa
		<voice name="ooo1">aaa1</voice>
		end
		</voice>
		end def</speak>`,
			want: []Part{
				&Text{Text: "aaa", Voice: "ooo", Speed: 1},
				&Text{Text: "aaa1", Voice: "ooo1", Speed: 1},
				&Text{Text: "end", Voice: "ooo", Speed: 1},
				&Text{Text: "end def", Voice: "aa", Speed: 1},
			}, wantErr: false},
		{name: "<voice> map", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf: func(s string) (string, error) { return "ooo.v1", nil },
			want: []Part{
				&Text{Text: "aaa", Voice: "ooo.v1", Speed: 1},
			}, wantErr: false},
		{name: "<voice> map fail", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf:   func(s string) (string, error) { return "", errors.New("no voice") },
			want: []Part{}, wantErr: true},
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
