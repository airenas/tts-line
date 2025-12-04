package audio

import (
	"bytes"
	"testing"
)

func Test_toInt16(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want int16
	}{
		{name: "zero", f: 0.0, want: 0},
		{name: "max", f: 32780, want: 32767},
		{name: "min", f: -32780, want: -32768},
		{name: "half max", f: 16383.2, want: 16383},
		{name: "half min", f: -16383.9, want: -16384},
		{name: "over max", f: 32767, want: 32767},
		{name: "over min", f: -32768, want: -32768},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInt16(tt.f)
			if got != tt.want {
				t.Errorf("toInt16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ChangeVolume(t *testing.T) {
	tests := []struct {
		name      string
		b         []byte
		volChange []*VolChange
		want      []byte
		wantErr   bool
	}{
		{name: "sil", b: []byte{0, 1, 2, 3, 4, 5, 6, 7}, volChange: []*VolChange{{From: 2, To: 6, Rate: 0}}, want: []byte{0, 1, 0, 0, 0, 0, 6, 7}, wantErr: false},
		{name: "no change", b: []byte{0, 1, 2, 3, 4, 5, 6, 7}, volChange: []*VolChange{}, want: []byte{0, 1, 2, 3, 4, 5, 6, 7}, wantErr: false},
		{name: "sil", b: []byte{0, 1, 2, 3, 4, 5, 6, 7}, volChange: []*VolChange{{From: 0, To: 2, Rate: 0}}, want: []byte{0, 0, 2, 3, 4, 5, 6, 7}, wantErr: false},
		{name: "initial rate", b: []byte{0, 1, 1, 0, 4, 5, 6, 7}, volChange: []*VolChange{{From: 2, To: 4, Rate: 0, StartRate: 10}}, want: []byte{0, 1, 10, 0, 4, 5, 6, 7}, wantErr: false},
		{name: "end rate", b: []byte{1, 0, 1, 0, 1, 0, 1, 0}, volChange: []*VolChange{{From: 2, To: 8, Rate: 0, EndRate: 10}}, want: []byte{1, 0, 0, 0, 0, 0, 10, 0}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ChangeVolume(t.Context(), tt.b, tt.volChange, 2)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("changeVolume() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("changeVolume() succeeded unexpectedly")
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("changeVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
