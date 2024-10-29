package utils

import "testing"

func TestErrTextTooLong_Error(t *testing.T) {
	type fields struct {
		Max int
		Len int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "name", fields: fields{Max: 10, Len: 15}, want: "text size too long, passed 15 chars, max 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ErrTextTooLong{
				Max: tt.fields.Max,
				Len: tt.fields.Len,
			}
			if got := r.Error(); got != tt.want {
				t.Errorf("ErrTextTooLong.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
