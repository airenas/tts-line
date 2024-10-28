package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_findLastConsequtiveAlign(t *testing.T) {
	type args struct {
		partlyAll []int
		oStrs     []string
		nStrs     []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "ok - the last", args: args{partlyAll: []int{0, 1, 2, 3, 4, 5}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "d", "e", "f"}}, want: 5, wantErr: false},
		{name: "ok - the last shifted", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "i", "d", "e", "f"}}, want: 5, wantErr: false},
		{name: "ok - not the last", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b", "c", "i", "d", "e", "fc"}}, want: 4, wantErr: false},
		{name: "fails match", args: args{partlyAll: []int{0, 1, 2, 4, 5, 6}, oStrs: []string{"a", "b", "c", "d", "e", "f"},
			nStrs: []string{"a", "b,", "c,", "i", "d", "e,", "f,"}}, want: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findLastConsequtiveAlign(tt.args.partlyAll, tt.args.oStrs, tt.args.nStrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("findLastConsequtiveAlign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findLastConsequtiveAlign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_align(t *testing.T) {
	type args struct {
		oStrs []string
		nStrs []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{name: "with shift", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: append([]int{0}, makeTestInt(6, 105)...), wantErr: false},
		{name: "simple", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "c", "d"}},
			want: []int{0, 1, 2, 3}, wantErr: false},
		{name: "insert", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "b1", "c", "d"}},
			want: []int{0, 1, 3, 4}, wantErr: false},
		{name: "skip", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "c", "d"}},
			want: []int{0, -1, 1, 2}, wantErr: false},
		{name: "long", args: args{oStrs: makeTestStr("a", 1000), nStrs: makeTestStr("a", 1000)},
			want: makeTestInt(0, 1000), wantErr: false},
		{name: "super long", args: args{oStrs: makeTestStr("a", 10000), nStrs: makeTestStr("a", 10000)},
			want: makeTestInt(0, 10000), wantErr: false},
		{name: "fails", args: args{oStrs: makeTestStr("a", 30), nStrs: makeTestStr("b", 30)},
			want: nil, wantErr: true},
		{name: "with shift", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: append([]int{0}, makeTestInt(6, 105)...), wantErr: false},
		{name: "with in middle", args: args{
			oStrs: append(makeTestStr("a", 100), makeTestStr("a", 100)...),
			nStrs: append(append(makeTestStr("a", 100), makeTestStr("c", 5)...), makeTestStr("a", 100)...)},
			want: append(makeTestInt(0, 100), makeTestInt(105, 205)...), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := align(tt.args.oStrs, tt.args.nStrs, 20)
			if (err != nil) != tt.wantErr {
				t.Errorf("align() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("align() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeTestStr(pr string, n int) []string {
	res := []string{}
	for i := 0; i < n; i++ {
		res = append(res, fmt.Sprintf("%s-%d", pr, i))
	}
	return res
}

func makeTestInt(from, n int) []int {
	res := []int{}
	for i := from; i < n; i++ {
		res = append(res, i)
	}
	return res
}
