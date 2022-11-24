package roller

import (
	"reflect"
	"testing"
)

func TestGetWodRollParams(t *testing.T) {
	type args struct {
		input []string
	}
	tests := []struct {
		name         string
		args         args
		wantParams   ThresholdParameters
		wantIsChance bool
		wantRollStr  string
		wantErr      bool
	}{
		// TODO: Add test cases.
		{
			"nothing_special",
			args{[]string{"5", "+", "7"}},
			ThresholdParameters{
				8,
				10,
			},
			false,
			"5 + 7",
			false,
		},
		{
			"9again_in_middle",
			args{[]string{"5", "+", "9again", "7"}},
			ThresholdParameters{8, 9},
			false,
			"5 + 7",
			false,
		},
		{
			"8again",
			args{[]string{"5", "+", "7", "8again"}},
			ThresholdParameters{8, 8},
			false,
			"5 + 7",
			false,
		},
		{
			"chance",
			args{[]string{"chance", "5", "+", "7"}},
			ThresholdParameters{8, 10},
			true,
			"5 + 7",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParams, gotIsChance, gotRollStr, err := GetWodRollParams(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWodRollParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("GetWodRollParams() gotParams = %v, want %v", gotParams, tt.wantParams)
			}
			if gotIsChance != tt.wantIsChance {
				t.Errorf("GetWodRollParams() gotIsChance = %v, want %v", gotIsChance, tt.wantIsChance)
			}
			if gotRollStr != tt.wantRollStr {
				t.Errorf("GetWodRollParams() gotRollStr = %v, want %v", gotRollStr, tt.wantRollStr)
			}
		})
	}
}
