package roller

import "testing"

func TestThresholdRoll_Repr(t1 *testing.T) {
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
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThresholdRoll{
				params: tt.fields.params,
				rolls:  tt.fields.rolls,
			}
			got, err := t.Repr()
			if (err != nil) != tt.wantErr {
				t1.Errorf("Repr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("Repr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThresholdRoll_Value(t1 *testing.T) {
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
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThresholdRoll{
				params: tt.fields.params,
				rolls:  tt.fields.rolls,
			}
			if got := t.Value(); got != tt.want {
				t1.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
