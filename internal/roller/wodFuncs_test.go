package roller

import (
	"reflect"
	"testing"
)

func TestNewGetWodRollParams(t *testing.T) {
	type args struct {
		isNineAgain  bool
		isEightAgain bool
	}
	tests := []struct {
		name string
		args args
		want ThresholdParameters
	}{
		{
			"default",
			args{
				false,
				false,
			},
			ThresholdParameters{passOn: 8, explodeOn: 10},
		},
		{
			"9again",
			args{
				true,
				false,
			},
			ThresholdParameters{passOn: 8, explodeOn: 9},
		},
		{
			"8again only",
			args{
				false,
				true,
			},
			ThresholdParameters{passOn: 8, explodeOn: 8},
		},
		{
			"both",
			args{
				true,
				true,
			},
			ThresholdParameters{8, 8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGetWodRollParams(tt.args.isNineAgain, tt.args.isEightAgain); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGetWodRollParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
