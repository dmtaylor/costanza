package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/dmtaylor/costanza/internal/roller"
)

const testSeed = 12345
const shortParseExpression = "1d6 + 2"
const longParseExpression = "(2d4 + 2 * 2) - (5d6 - 5) / 3 + (10d10 + 20d4 * 1d4)"

func TestDNotationParser_DoParse(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult *DNotationResult
		expectedErr    error
	}{
		// TODO add tests
		{
			"simple_arithmetic_test",
			"4 + 5",
			&DNotationResult{
				Value:    9,
				StrValue: "4 + 5",
			},
			nil,
		},
		{
			"simple_roll_test",
			"3d6",
			&DNotationResult{
				Value:    8,
				StrValue: "[3 + 3 + 2]",
			},
			nil,
		},
		{
			"complex_roll_test",
			"5d10 + 2 * (2d12 - 3d4)",
			&DNotationResult{
				Value:    69,
				StrValue: "[9 + 9 + 6 + 9 + 8] + 2 * ( [11 + 11] - [2 + 2 + 4] )",
			},
			nil,
		},
		{
			name:  "simple_lexing_error",
			input: "5 + alphachars",
			expectedErr: fmt.Errorf("failed to parse string: %w", &participle.ParseError{
				Msg: "invalid input text \"alphachars\"",
				Pos: lexer.Position{Line: 1, Column: 5},
			}),
		},
		{
			name:  "bad_parsing",
			input: "6 + 5 + ",
			expectedErr: fmt.Errorf("failed to parse string: %w", &participle.UnexpectedTokenError{
				Expect:     "Term",
				Unexpected: lexer.EOFToken(lexer.Position{Line: 1, Column: 9}),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewDNotationParser()
			if err != nil {
				t.Errorf("error when building parser: %s", err.Error())
				return
			}
			parser.roller = roller.NewTestBaseRoller(testSeed)
			result, err := parser.DoParse(tt.input)
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("DoParse() error = %+v", err)
					return
				}
				if !reflect.DeepEqual(result, tt.expectedResult) {
					t.Errorf("DoParse() got = %+v, want %+v", result, tt.expectedResult)
					return
				}
			} else {
				if result != nil {
					t.Errorf("DoParse() error expected nil, got res = %+v", result)
					return
				}
				if err.Error() != tt.expectedErr.Error() {
					t.Errorf("DoParse() got = (%T)%+v, want = (%T)%+v", err, err, tt.expectedErr, tt.expectedErr)
					return
				}
			}
		})
	}

}

func TestNewDNotationParser(t *testing.T) {
	got, err := NewDNotationParser()
	if err != nil {
		t.Errorf("error when building parser: %s", err.Error())
		return
	}
	if got == nil {
		t.Errorf("nil DNotationParser")
		return
	}
}

// TODO add benchmark tests

func BenchmarkDNotationParser_DoParseShort(b *testing.B) {
	parser, err := NewDNotationParser()
	if err != nil {
		b.Errorf("failed to build parser: %s", err.Error())
		return
	}
	parser.roller = roller.NewTestBaseRoller(testSeed)
	for i := 0; i < b.N; i++ {
		_, err = parser.DoParse(shortParseExpression)
		if err != nil {
			b.Errorf("failure running expression: %v", err)
			return
		}
	}
}

func BenchmarkDNotationParser_DoParseLong(b *testing.B) {
	parser, err := NewDNotationParser()
	if err != nil {
		b.Errorf("failed to build parser: %s", err.Error())
		return
	}
	parser.roller = roller.NewTestBaseRoller(testSeed)
	for i := 0; i < b.N; i++ {
		_, err = parser.DoParse(longParseExpression)
		if err != nil {
			b.Errorf("failure running expression: %v", err)
			return
		}
	}
}
