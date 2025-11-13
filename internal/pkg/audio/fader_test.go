package audio

import (
	"testing"

	"github.com/airenas/tts-line/internal/pkg/utils"
)

func TestFader_at(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		pos  int
		want float64
	}{
		{name: "before", pos: -1, want: 0.0},
		{name: "start", pos: 0, want: 0.0066928509242848},
		{name: "middle", pos: 550, want: 0.4971655632430104},
		{name: "end", pos: 1101, want: 0.9932160895691214},
		{name: "after", pos: 1105, want: 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := defaultFader
			got := f.at(tt.pos)
			if utils.Float64Equals(got, tt.want) == false {
				t.Errorf("at() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFader_Fade(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		pos  int
		bl   int
		want float64
	}{
		{name: "start", pos: 0, bl: 5000, want: 0.0066928509242848},
		{name: "start 10", pos: 10, bl: 5000, want: 0.007323641200864822},
		{name: "start 1000", pos: 1000, bl: 5000, want: 0.9832142177368054},
		{name: "start 1120", pos: 1120, bl: 5000, want: 1.0},
		{name: "middle", pos: 2000, bl: 5000, want: 1.0},
		{name: "middle end", pos: 3000, bl: 5000, want: 1.0},
		{name: "end 1000", pos: 4000, bl: 5000, want: 0.9830638634560174},
		{name: "end", pos: 4999, bl: 5000, want: 0.0066928509242848554},
		{name: "after2", pos: 6000, bl: 5000, want: 0.0},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := defaultFader
			got := f.Fade(tt.pos, tt.bl)
			if utils.Float64Equals(got, tt.want) == false {
				t.Errorf("Fade() = %v, want %v", got, tt.want)
			}
		})
	}
}
