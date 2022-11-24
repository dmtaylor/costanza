package util

import (
	"reflect"
	"testing"
)

func TestIntSliceToStr(t *testing.T) {
	type args struct {
		in []int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"main",
			args{[]int{5, 3, 1, 10}},
			[]string{"5", "3", "1", "10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntSliceToStr(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IntSliceToStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
