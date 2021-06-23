package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/dmtaylor/costanza/internal/roller"
	"github.com/pkg/errors"
)

type Operator int

type BasicParser struct {
	roller *roller.BaseRoller
	parser *participle.Parser
}

type DNotationResult struct {
	Value    int
	StrValue string
}

const (
	OpMul  Operator = iota // '*'
	OpDiv                  // '/'
	OpAdd                  // '+'
	OpSub                  // '-'
	OpRoll                 // 'd'
)

var operatorMap = map[string]Operator{"*": OpMul, "/": OpDiv, "+": OpAdd, "-": OpSub, "d": OpRoll}
var reverseOperatorMap = map[Operator]string{OpMul: "*", OpDiv: "/", OpAdd: "+", OpSub: "-", OpRoll: "d"}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

type Value struct {
	Number        int         `  @(Number)`
	SubExpression *Expression `| "(" @@ ")"`
}

type DFactor struct {
	Count *Value `  @@`
	//Operator Operator `  "d"`
	Sides *Value `  "d"@@`
	Value *Value `| @@`
}

type OpFactor struct {
	Operator Operator `@("*" | "/")`
	DFactor  *DFactor `@@`
}

type Term struct {
	Left  *DFactor    `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

type Expression struct {
	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

func (o Operator) Eval(l, r int) (*DNotationResult, error) {
	var value int
	switch o {
	case OpAdd:
		value = l + r
	case OpSub:
		value = l - r
	case OpMul:
		value = l * r
	case OpDiv:
		value = l / r
	default:
		return nil, errors.Errorf("invalid operator: %s", reverseOperatorMap[o])
	}
	return &DNotationResult{
		Value:    value,
		StrValue: fmt.Sprintf(" %s ", reverseOperatorMap[o]),
	}, nil
}

func (v *Value) Eval(roller *roller.BaseRoller) (*DNotationResult, error) {
	if v.SubExpression != nil {
		subRes, err := v.SubExpression.Eval(roller)
		if err != nil {
			return nil, err
		}
		return &DNotationResult{
			Value:    subRes.Value,
			StrValue: fmt.Sprintf("( %s )", subRes.StrValue),
		}, nil
	} else {
		return &DNotationResult{
			Value:    v.Number,
			StrValue: strconv.Itoa(v.Number),
		}, nil
	}
}

func (d *DFactor) Eval(roller *roller.BaseRoller) (*DNotationResult, error) {
	if d.Value != nil {
		return d.Value.Eval(roller)
	}
	leftRes, err := d.Count.Eval(roller)
	if err != nil {
		return nil, err
	}
	rightRes, err := d.Sides.Eval(roller)
	if err != nil {
		return nil, err
	}
	rollRes := roller.DoRoll(leftRes.Value, rightRes.Value)
	rollStr, err := rollRes.Repr()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build roll result string")
	}
	return &DNotationResult{
		rollRes.Sum(),
		rollStr,
	}, nil

}

func (t *Term) Eval(roller *roller.BaseRoller) (*DNotationResult, error) {
	l, err := t.Left.Eval(roller)
	if err != nil {
		return nil, err
	}
	accum := l.Value
	strAccum := new(strings.Builder)
	_, err = strAccum.WriteString(l.StrValue)
	if err != nil {
		return nil, err
	}
	for _, r := range t.Right {
		rightFactor, err := r.DFactor.Eval(roller)
		if err != nil {
			return nil, err
		}
		opRes, err := r.Operator.Eval(accum, rightFactor.Value)
		if err != nil {
			return nil, err
		}
		accum = opRes.Value
		_, err = strAccum.WriteString(opRes.StrValue)
		if err != nil {
			return nil, err
		}
		_, err = strAccum.WriteString(rightFactor.StrValue)
		if err != nil {
			return nil, err
		}
	}
	return &DNotationResult{
		accum,
		strAccum.String(),
	}, nil
}

func (e *Expression) Eval(roller *roller.BaseRoller) (*DNotationResult, error) {

	l, err := e.Left.Eval(roller)
	if err != nil {
		return nil, err
	}
	accum := l.Value
	strAccum := new(strings.Builder)
	_, err = strAccum.WriteString(l.StrValue)
	if err != nil {
		return nil, err
	}

	for _, r := range e.Right {
		rightTerm, err := r.Term.Eval(roller)
		if err != nil {
			return nil, err
		}
		opRes, err := r.Operator.Eval(accum, rightTerm.Value)
		if err != nil {
			return nil, err
		}
		accum = opRes.Value
		_, err = strAccum.WriteString(opRes.StrValue)
		if err != nil {
			return nil, err
		}
		_, err = strAccum.WriteString(rightTerm.StrValue)
		if err != nil {
			return nil, err
		}

	}
	return &DNotationResult{
		accum,
		strAccum.String(),
	}, nil
}

func getLexer() (*stateful.Definition, error) {
	return stateful.NewSimple([]stateful.Rule{
		{"Operator", `[*/+=d()]`, nil},
		{"Number", `\d+`, nil},
		{"whitespace", `\s+`, nil},
	})

}

func NewBasicParser() (*BasicParser, error) {
	lexer, err := getLexer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build lexer")
	}
	parser, err := participle.Build(&Expression{}, participle.Lexer(lexer))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build parser")
	}
	return &BasicParser{
		roller: roller.New(),
		parser: parser,
	}, nil
}

func (p *BasicParser) GetEBNF() string {
	return p.parser.String()
}

func (p *BasicParser) DoParse(input string) (*DNotationResult, error) {
	expr := &Expression{}
	err := p.parser.ParseString("", input, expr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse string")
	}
	return expr.Eval(p.roller)
}
