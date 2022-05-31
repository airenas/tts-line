package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFileName(t *testing.T) {
	type args struct {
		b   string
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Simple", args: args{b: "in", key: "aa.wav"}, want: "in/aa.wav"},
		{name: "Strips key", args: args{b: "in", key: "olia/aa.wav"}, want: "in/aa.wav"},
		{name: "Strips key", args: args{b: "in", key: "../aa.wav"}, want: "in/aa.wav"},
		{name: "Leaves base", args: args{b: "../in/path", key: "aa.wav"}, want: "../in/path/aa.wav"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFileName(tt.args.b, tt.args.key); got != tt.want {
				t.Errorf("getFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLoader(t *testing.T) {
	l, err := NewLoader("")
	assert.Nil(t, l)
	assert.NotNil(t, err)
	l, err = NewLoader("./")
	assert.NotNil(t, l)
	assert.Nil(t, err)
	l, err = NewLoader("loader.go")
	assert.Nil(t, l)
	assert.NotNil(t, err)
	l, err = NewLoader("xxx/eee/eee")
	assert.Nil(t, l)
	assert.NotNil(t, err)
}
