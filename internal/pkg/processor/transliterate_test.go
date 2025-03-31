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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareWords(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("compareWords() = %v, want %v", got, tt.want)
			}
		})
	}
}
