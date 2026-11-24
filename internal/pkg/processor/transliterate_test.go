package processor

import "testing"

func Test_compareWords(t *testing.T) {
	type args struct {
		old string
		new string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "same", args: args{old: "hello", new: "hello"}, want: true},
		{name: "different", args: args{old: "hello", new: "world"}, want: false},
		{name: "different with apostrophe", args: args{old: "hello", new: "he'llo"}, want: false},
		{name: "same with apostrophe", args: args{old: "he’llo", new: "he'llo"}, want: true},
		{name: "same with apostrophe 2", args: args{old: "he`llo", new: "he'llo"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareWords(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("compareWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkUpper(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want string
	}{
		{name: "both uppercase", s1: "Hello", s2: "Hello", want: "Hello"},
		{name: "need change", s1: "hello", s2: "Hello", want: "Hello"},
		{name: "need change non asci", s1: "ąello", s2: "Ąello", want: "Ąello"},
		{name: "no change", s1: "hello", s2: "hello", want: "hello"},
		{name: "not same letters", s1: "hello", s2: "World", want: "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkUpper(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("checkUpper() = %v, want %v", got, tt.want)
			}
		})
	}
}
