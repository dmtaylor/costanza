package roller

import "testing"

func TestBaseRoll_Repr(t *testing.T) {
	tests := []struct {
		name    string
		r       BaseRoll
		want    string
		wantErr bool
	}{
		{"single", BaseRoll{5}, "[5]", false},
		{"multiple", BaseRoll{5, 8, 10}, "[5 + 8 + 10]", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Repr()
			if (err != nil) != tt.wantErr {
				t.Errorf("Repr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Repr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseRoll_Sum(t *testing.T) {
	tests := []struct {
		name string
		r    BaseRoll
		want int
	}{
		{"single", BaseRoll{5}, 5},
		{"multiple", BaseRoll{5, 8, 10}, 23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Sum(); got != tt.want {
				t.Errorf("Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseRoll_Value(t *testing.T) {
	tests := []struct {
		name string
		r    BaseRoll
		want int
	}{
		{"single", BaseRoll{5}, 5},
		{"multiple", BaseRoll{5, 8, 10}, 23},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
