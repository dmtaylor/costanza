package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidQuoteTypeError_Error(t *testing.T) {
	tests := []struct {
		name string
		i    InvalidQuoteTypeError
		want string
	}{
		{
			"baseline",
			InvalidQuoteTypeError("flugelhorn"),
			"invalid quote type flugelhorn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.i, tt.want)
		})
	}
}
