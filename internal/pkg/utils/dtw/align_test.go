package dtw

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/utils"
)

func Test_Align(t *testing.T) {
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
		{name: "real3", args: args{oStrs: []string{"(1988", "m.", "surinkęs", "25", "proc.", "balsų),",
			"bjūkenenas", "(1996", "m.", "surinkęs", "23", "proc.", "balsų)",
			"ir", "forbzas", "(2000", "m.", "surinkęs", "31", "proc.", "balsų)"},
			nStrs: []string{"(tūkstantis", "devyni", "šimtai", "aštuoniasdešimt", "aštuntieji", "metai", "surinkęs", "dvidešimt", "penki", "procentai", "balsų),",
				"bjūkenenas", "(tūkstantis", "devyni", "šimtai", "devyniasdešimt", "šeštieji", "metai", "surinkęs", "dvidešimt", "trys", "procentai", "balsų)",
				"ir", "forbzas", "(du", "tūkstantieji", "metai", "surinkęs", "trisdešimt", "vienas", "procentas", "balsų)"}},
			wantErr: false, want: []int{0, 1, 6, 7, 8, 10, 11, 12, 13, 18, 19, 20, 22, 23, 24, 25, 26, 28, 29, 30, 32}},
		{name: "real", args: args{oStrs: []string{"122", "-", "122", "bandymai."},
			nStrs: []string{"šimtas", "dvidešimt", "du", "-", "šimtas", "dvidešimt", "du", "bandymai."}},
			want: []int{0, 3, 4, 7}, wantErr: false},
		{name: "real 2", args: args{oStrs: []string{"122", "bandymai."},
			nStrs: []string{"šimtas", "dvidešimt", "du", "bandymai."}},
			want: []int{0, 3}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Align(t.Context(), tt.args.oStrs, tt.args.nStrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("align() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// for i, v := range got {
				// 	fmt.Printf("got[%d] = %v, want %v\n", i, v, tt.want[i])
				// }
				t.Errorf("align() = %v, want %v", got, tt.want)
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
		{name: "insert 1", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"b1", "a", "b", "c", "d"}},
			want: []int{1, 2, 3, 4}, wantErr: false},
		{name: "same in a row small", args: args{oStrs: []string{"b", "c", "c", "k"}, nStrs: []string{"b", "c", "c", "e", "k"}},
			want: []int{0, 1, 2, 4}, wantErr: false},
		{name: "same in a row", args: args{oStrs: []string{"a", "b", "c", "c", "d", "k"}, nStrs: []string{"a", "b", "c", "c", "d", "e", "k"}},
			want: []int{0, 1, 2, 3, 4, 6}, wantErr: false},
		{name: "three in a row ", args: args{oStrs: []string{"a", "b", "c", "c", "c", "d", "k"}, nStrs: []string{"a", "b", "c", "c", "c", "d", "k"}},
			want: []int{0, 1, 2, 3, 4, 5, 6}, wantErr: false},
		{name: "empty", args: args{oStrs: []string{}, nStrs: []string{}},
			want: []int{}, wantErr: false},
		{name: "simple", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "c", "d"}},
			want: []int{0, 1, 2, 3}, wantErr: false},
		{name: "insert middle", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "b", "b1", "c", "d"}},
			want: []int{0, 1, 3, 4}, wantErr: false},
		{name: "with shift", args: args{oStrs: makeTestStr("a", 10),
			nStrs: append(makeTestStr("c", 3), makeTestStr("a", 10)...)},
			want: makeTestInt(3, 13), wantErr: false},
		{name: "insert", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"b1", "b1", "b1", "a", "b", "c", "d"}},
			want: []int{3, 4, 5, 6}, wantErr: false},
		{name: "skip", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"a", "c", "d"}},
			want: []int{0, -1, 1, 2}, wantErr: false},
		{name: "with shift 5 100", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: append([]int{}, makeTestInt(5, 105)...), wantErr: false},
		{name: "skip first", args: args{oStrs: []string{"a", "b", "c", "d"}, nStrs: []string{"b", "c", "d"}},
			want: []int{-1, 0, 1, 2}, wantErr: false},
		{name: "skip first", args: args{oStrs: []string{"a", "a", "b", "c", "d"}, nStrs: []string{"b", "c", "d"}},
			want: []int{-1, -1, 0, 1, 2}, wantErr: false},
		{name: "skip first 2", args: args{oStrs: []string{"b", "c", "d"}, nStrs: []string{"a", "a", "b", "c", "d"}},
			want: []int{2, 3, 4}, wantErr: false},
		{name: "long", args: args{oStrs: makeTestStr("a", 1000), nStrs: makeTestStr("a", 1000)},
			want: makeTestInt(0, 1000), wantErr: false},
		{name: "super long", args: args{oStrs: makeTestStr("a", 10000), nStrs: makeTestStr("a", 10000)},
			want: makeTestInt(0, 10000), wantErr: false},
		{name: "fails", args: args{oStrs: makeTestStr("a", 30), nStrs: makeTestStr("b", 30)},
			want: makeTestInt(0, 30), wantErr: false},
		{name: "with shift 5", args: args{oStrs: makeTestStr("a", 100),
			nStrs: append(makeTestStr("c", 5), makeTestStr("a", 100)...)},
			want: makeTestInt(5, 105), wantErr: false},
		{name: "with in middle", args: args{
			oStrs: append(makeTestStr("a", 100), makeTestStr("a", 100)...),
			nStrs: append(append(makeTestStr("a", 100), makeTestStr("c", 5)...), makeTestStr("a", 100)...)},
			want: append(makeTestInt(0, 100), makeTestInt(105, 205)...), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := alignBanded(tt.args.oStrs, tt.args.nStrs, 14)
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

func BenchmarkAlignInt(b *testing.B) {
	type scenario struct {
		name      string
		origLen   int
		insert    int
		bandWidth int
	}

	scenarios := []scenario{
		{"10 x 20", 10, 10, 20},
		{"100 x 20", 100, 10, 20},
		{"500 x 20", 500, 10, 20},
		{"500 x 40", 500, 20, 40},
		{"2000 x 20", 2000, 10, 20},
		{"2000 x 40", 2000, 20, 40},
		{"5000 x 40", 5000, 20, 40},
		{"5000 x 100", 5000, 50, 100},
		{"5000 x 200", 5000, 100, 200},
	}

	for _, s := range scenarios {
		// Precompute inputs
		orig := append(makeTestStr("a", s.origLen/2), makeTestStr("a", s.origLen/2)...)
		pred := append(makeTestStr("a", s.origLen/2), append(makeTestStr("b", s.insert), makeTestStr("a", s.origLen/2)...)...)

		// Use b.Run to create sub-benchmarks
		b.Run(s.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := alignBanded(orig, pred, s.bandWidth)
				if err != nil {
					b.Errorf("alignBanded() error = %v", err)
				}
			}
		})
	}
}

func BenchmarkAlign(b *testing.B) {
	type scenario struct {
		name    string
		origLen int
		insert  int
	}

	scenarios := []scenario{
		{"10 ", 10, 20},
		{"100", 100, 20},
		{"500 ", 500, 30},
		{"500 ", 500, 30},
		{"2000 ", 2000, 30},
		{"2000 i30", 2000, 30},
		{"5000 i40", 5000, 40},
		{"5000 i50", 5000, 50},
	}

	for _, s := range scenarios {
		// Precompute inputs
		orig := append(makeTestStr("a", s.origLen/2), makeTestStr("a", s.origLen/2)...)
		pred := append(makeTestStr("a", s.origLen/2), append(makeTestStr("b", 10), makeTestStr("a", s.origLen/2)...)...)

		// Use b.Run to create sub-benchmarks
		b.Run(s.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := Align(b.Context(), orig, pred)
				if err != nil {
					b.Errorf("Align() error = %v", err)
				}
			}
		})
	}
}

func BenchmarkAlignOld(b *testing.B) {
	type scenario struct {
		name    string
		origLen int
		predLen int
	}
	scenarios := []scenario{
		{"10 ", 10, 20},
		{"100", 100, 20},
		{"500 ", 500, 30},
		{"500 ", 500, 30},
		{"2000 ", 2000, 30},
		{"2000 i30", 2000, 30},
		{"5000 i40", 5000, 40},
		{"5000 i50", 5000, 50},
	}

	for _, s := range scenarios {
		// Precompute inputs
		orig := append(makeTestStr("a", s.origLen/2), makeTestStr("a", s.origLen/2)...)
		pred := append(makeTestStr("a", s.origLen/2), append(makeTestStr("b", 10), makeTestStr("a", s.origLen/2)...)...)

		// Use b.Run to create sub-benchmarks
		b.Run(s.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = utils.Align(orig, pred)
			}
		})
	}
}
