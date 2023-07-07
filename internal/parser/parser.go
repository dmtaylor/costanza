package parser

// RollParser interface to support parsing a d-notation roll & getting a result
type RollParser interface {
	DoParse(string) (*DNotationResult, error)
}

type DNotationResult struct {
	Value    int
	StrValue string
}
