package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocessRoll(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"no_change",
			"1d20 + 4",
			"1d20 + 4",
		},
		{
			"bare_roll_beginning",
			"d20 + 1d20",
			"1d20 + 1d20",
		},
		{
			"mulitiple_bare_rolls",
			"1d20 + d5+d8",
			"1d20 + 1d5+1d8",
		},
		{
			"empty_string",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PreprocessRoll(tt.input)
			assert.Equal(t, tt.want, got, "mismatched preprocessing")
		})
	}
}

func ExamplePreprocessRoll() {
	s := PreprocessRoll("d20 + d8 + 1d4")
	fmt.Println(s)
	// Output: 1d20 + 1d8 + 1d4
}
