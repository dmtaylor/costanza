package roller

import (
	"strings"

	"github.com/dmtaylor/costanza/internal/util"
)

type Roll interface {
	Repr() string
	Value() int
}

type BaseRoll []int

func (r *BaseRoll) Repr() (string, error) {
	builder := strings.Builder{}
	_, err := builder.WriteString("[")
	if err != nil {
		return "", err
	}
	_, err = builder.WriteString(strings.Join(util.IntSliceToStr([]int(*r)), " + "))
	if err != nil {
		return "", nil
	}
	_, err = builder.WriteString("]")
	if err != nil {
		return "", nil
	}

	return builder.String(), nil
}

func (r *BaseRoll) Sum() int {
	sum := 0
	for _, value := range *r {
		sum += value
	}
	return sum
}

func (r *BaseRoll) Value() int {
	return r.Sum()
}
