package ssml

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &Text{Voice: "aa", Speed: 1}
			got, err := Parse(strings.NewReader(tt.xml), def)
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
