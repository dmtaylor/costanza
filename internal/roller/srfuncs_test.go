package roller

import (
	"reflect"
	"testing"
)

func TestGetGlitchStatus(t *testing.T) {
	type args struct {
		roll ThresholdRoll
	}
	params := GetSrParams()
	tests := []struct {
		name string
		args args
		want SrGlitchStatus
	}{
		{
			"noGlitch",
			args{ThresholdRoll{
				params,
				[]singleThresholdRoll{
					{3, false},
					{5, false},
					{1, false},
				},
			}},
			SrNoGlitch,
		},
		{
			"glitch",
			args{ThresholdRoll{
				params,
				[]singleThresholdRoll{
					{1, false},
					{6, false},
					{1, true},
					{1, false},
				},
			}},
			SrGlitch,
		},
		{
			"critGlitch",
			args{ThresholdRoll{
				params,
				[]singleThresholdRoll{
					{1, false},
					{2, false},
					{1, false},
				},
			}},
			SrCritGlitch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGlitchStatus(tt.args.roll); got != tt.want {
				t.Errorf("GetGlitchStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSrParams(t *testing.T) {
	tests := []struct {
		name string
		want ThresholdParameters
	}{
		{"main", ThresholdParameters{5, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSrParams(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSrParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isGlitch(t *testing.T) {
	type args struct {
		roll ThresholdRoll
	}
	params := GetSrParams()
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"notGlitch",
			args{ThresholdRoll{
				params,
				[]singleThresholdRoll{
					{4, false},
					{5, false},
					{3, false},
					{1, false},
				},
			}},
			false,
		},
		{
			"isGlitch",
			args{ThresholdRoll{
				params,
				[]singleThresholdRoll{
					{1, false},
					{6, false},
					{1, true},
					{1, false},
				},
			}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGlitch(tt.args.roll); got != tt.want {
				t.Errorf("isGlitch() = %v, want %v", got, tt.want)
			}
		})
	}
}
