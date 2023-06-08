package util

import (
	"testing"
	"time"
)

func TestGetLastMonth(t *testing.T) {
	type args struct {
		now time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "example",
			args: args{now: time.Date(2022, 05, 01, 0, 1, 1, 0, time.UTC)},
			want: "2022-04",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLastMonth(tt.args.now); got != tt.want {
				t.Errorf("GetLastMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
