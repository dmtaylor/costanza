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

type DNotationParser struct {
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

type OpDValue struct {
	Operator Operator `"d"`
	Value    *Value   `@@`
}

type Factor struct {
	Left  *Value      `@@`
	Right []*OpDValue `@@*`
}

type OpFactor struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

type Term struct {
	Left  *Factor     `@@`
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

func (d *Factor) Eval(baseRoller *roller.BaseRoller) (*DNotationResult, error) {
	leftRes, err := d.Left.Eval(baseRoller)
	if err != nil {
		return nil, err
	}
	nrolls := leftRes.Value
	strVal := leftRes.StrValue
	for _, r := range d.Right {
		rightRes, err := r.Value.Eval(baseRoller)
		if err != nil {
			return nil, err
		}
		rollRes := baseRoller.DoRoll(nrolls, rightRes.Value)
		nrolls = rollRes.Sum()
		strVal, err = rollRes.Repr()
		if err != nil {
			return nil, err
		}
	}
	return &DNotationResult{
		nrolls,
		strVal,
	}, nil
}

func (t *Term) Eval(baseRoller *roller.BaseRoller) (*DNotationResult, error) {
	l, err := t.Left.Eval(baseRoller)
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
		rightFactor, err := r.Factor.Eval(baseRoller)
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

func (e *Expression) Eval(baseRoller *roller.BaseRoller) (*DNotationResult, error) {

	l, err := e.Left.Eval(baseRoller)
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
		rightTerm, err := r.Term.Eval(baseRoller)
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

func NewDNotationParser() (*DNotationParser, error) {
	lexer, err := getLexer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build lexer")
	}
	parser, err := participle.Build(&Expression{}, participle.Lexer(lexer))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build parser")
	}
	return &DNotationParser{
		roller: roller.New(),
		parser: parser,
	}, nil
}

func (p *DNotationParser) GetEBNF() string {
	return p.parser.String()
}

func (p *DNotationParser) DoParse(input string) (*DNotationResult, error) {
	expr := &Expression{}
	err := p.parser.ParseString("", input, expr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse string")
	}
	return expr.Eval(p.roller)
}
