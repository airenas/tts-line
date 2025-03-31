package transcription

import (
	"reflect"
	"testing"
)

func TestTrimAccent(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"no accent", args{"abc"}, "abc"},
		{"accent", args{"ą3"}, "ą"},
		{"accent", args{"ar3-ka"}, "ar-ka"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimAccent(tt.args.s); got != tt.want {
				t.Errorf("TrimAccent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want *Data
	}{
		{"empty", args{""}, &Data{Word: "", Sylls: "", Transcription: ""}},
		{"no accent", args{"abc"}, &Data{Word: "abc", Sylls: "abc", Transcription: "abc"}},
		{"accent", args{"abc3-da"}, &Data{Word: "abcda", Sylls: "abc-da", Transcription: "abc3da"}},
		{"accent", args{"Q-lium-pi-ja3-kQs"}, &Data{Word: "oliumpijakos", Sylls: "o-lium-pi-ja-kos",
			Transcription: "Oliumpija3kOs"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
