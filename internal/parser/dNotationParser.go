package parser

type Operator int

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

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

type Value struct {
	Number        int         `  @(Int)`
	SubExpression *Expression `| "(" @@ ")"`
}

type DFactor struct {
	Count *Value `  @@`
	Sides *Value `  "d" @@`
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
