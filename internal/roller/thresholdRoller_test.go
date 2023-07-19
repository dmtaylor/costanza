package roller

import (
	"math"
	"reflect"
	"testing"
)

func TestThresholdRoll_String(t *testing.T) {
	type fields struct {
		params ThresholdParameters
		rolls  []singleThresholdRoll
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"baseline",
			fields{
				ThresholdParameters{
					5,
					6,
				},
				[]singleThresholdRoll{
					{1, false},
					{2, false},
					{3, false},
				},
			},
			"1 2 3",
			false,
		},
		{
			"explosions_basic",
			fields{
				ThresholdParameters{5, 6},
				[]singleThresholdRoll{
					{1, false},
					{6, false},
					{3, true},
					{5, false},
				},
			},
			"1 6 (3) 5",
			false,
		},
		{
			"explosions_chained",
			fields{
				ThresholdParameters{5, 6},
				[]singleThresholdRoll{
					{1, false},
					{6, false},
					{6, true},
					{6, true},
					{3, true},
					{5, false},
				},
			},
			"1 6 (6) (6) (3) 5",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roll := &ThresholdRoll{
				params: tt.fields.params,
				rolls:  tt.fields.rolls,
			}
			got, err := roll.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThresholdRoll_Value(t *testing.T) {
	type fields struct {
		params ThresholdParameters
		rolls  []singleThresholdRoll
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"baseline",
			fields{
				ThresholdParameters{5, 6},
				[]singleThresholdRoll{
					{1, false},
					{5, false},
					{4, false},
				},
			},
			1,
		},
		{
			"explosions",
			fields{
				ThresholdParameters{5, 6},
				[]singleThresholdRoll{
					{1, false},
					{6, false},
					{6, true},
					{6, true},
					{3, true},
					{5, false},
				},
			},
			4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := &ThresholdRoll{
				params: tt.fields.params,
				rolls:  tt.fields.rolls,
			}
			if got := roller.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThresholdRoller_DoThresholdRoll(t *testing.T) {
	var testSeed uint64 = 1111111
	type args struct {
		count  int
		sides  int
		params ThresholdParameters
	}
	tests := []struct {
		name           string
		args           args
		want           ThresholdRoll
		wantErr        bool
		expectedValue  int
		expectedString string
	}{
		{
			"simple_threshold",
			args{
				2,
				10,
				ThresholdParameters{passOn: 8, explodeOn: math.MaxInt},
			},
			ThresholdRoll{
				params: ThresholdParameters{passOn: 8, explodeOn: math.MaxInt},
				rolls: []singleThresholdRoll{
					{value: 3},
					{value: 9},
				},
			},
			false,
			1,
			"3 9",
		},
		{
			"exploding_threshold",
			args{
				5,
				10,
				ThresholdParameters{passOn: 8, explodeOn: 9},
			},
			ThresholdRoll{
				params: ThresholdParameters{passOn: 8, explodeOn: 9},
				rolls: []singleThresholdRoll{
					{value: 3},
					{value: 9},
					{value: 2, isExplode: true},
					{value: 8},
					{value: 10},
					{value: 9, isExplode: true},
					{value: 2, isExplode: true},
					{value: 9},
					{value: 2, isExplode: true},
				},
			},
			false,
			5,
			"3 9 (2) 8 10 (9) (2) 9 (2)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := &ThresholdRoller{
				baseRoller: NewTestBaseRoller(testSeed),
			}
			got, err := roller.DoThresholdRoll(tt.args.count, tt.args.sides, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DoThresholdRoll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoThresholdRoll() got = %v, want %v", got, tt.want)
			}
			if tt.expectedValue != got.Value() {
				t.Errorf("DoThresholdRoll() value got = %v, want %v", got.Value(), tt.expectedValue)
			}
			gotString, err := got.String()
			if err != nil {
				t.Errorf("got error with string representation %v", err)
				return
			}
			if gotString != tt.expectedString {
				t.Errorf("DoThresholdRoll() string value got = %v, want %v", gotString, tt.expectedString)
			}
		})
	}
}
