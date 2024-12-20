package roller

import (
	"math/rand/v2"
	"sync"
	"testing"
)

func TestBaseRoller_getRoll(t *testing.T) {
	src := rand.NewPCG(1, 2)
	type fields struct {
		rng  *rand.Rand
		lock *sync.Mutex
	}
	type args struct {
		base int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		lowerBound int
		upperBound int
	}{
		{
			"basic",
			fields{rand.New(src), &sync.Mutex{}},
			args{6},
			1,
			6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BaseRoller{
				rng:  tt.fields.rng,
				lock: tt.fields.lock,
			}
			if got := r.getRoll(tt.args.base); got < tt.lowerBound || got > tt.upperBound {
				t.Errorf("getRoll() = %v, want %v <= %v <= %v", got, tt.lowerBound, got, tt.lowerBound)
			}
		})
	}
}

func TestBaseRoller_DoRoll(t *testing.T) {
	// src := &rand.PCGSource{}
	// src.Seed(rngSeed)
	src := rand.NewPCG(3, 4)
	type fields struct {
		rng  *rand.Rand
		lock *sync.Mutex
	}
	type args struct {
		num  int
		base int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		lowerBound int
		upperBound int
		rollCount  int
	}{
		{
			"basic_bounds",
			fields{rand.New(src), &sync.Mutex{}},
			args{5, 12},
			1,
			12,
			5,
		},
		{
			"large_roll_count",
			fields{rand.New(src), &sync.Mutex{}},
			args{300, 6},
			1,
			6,
			300,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BaseRoller{
				rng:  tt.fields.rng,
				lock: tt.fields.lock,
			}
			results := r.DoRoll(tt.args.num, tt.args.base)
			if len(results) != tt.rollCount {
				t.Errorf("DoRoll() = %v,  got %v results, expected %v", results, len(results), tt.rollCount)
			}
			for _, res := range results {
				if res < tt.lowerBound || res > tt.upperBound {
					t.Errorf("DoRoll() = %v, want %v <= %v <= %v", results, tt.lowerBound, res, tt.upperBound)
				}
			}
		})
	}
}

func BenchmarkBaseRoller_DoRoll(b *testing.B) {
	// src := &rand.PCGSource{}
	// src.Seed(rngSeed)
	src := rand.NewPCG(5, 6)
	r := &BaseRoller{
		rng:  rand.New(src),
		lock: &sync.Mutex{},
	}
	rollcount := 5
	sides := 20

	for i := 0; i < b.N; i++ {
		_ = r.DoRoll(rollcount, sides)
	}

}
