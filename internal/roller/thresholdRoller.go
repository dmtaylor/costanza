package roller

import (
	"fmt"
	"strings"
)

type singleThresholdRoll struct {
	value     int
	isExplode bool
}

type ThresholdParameters struct {
	passOn    int
	explodeOn int
}

type ThresholdRoll struct {
	params ThresholdParameters
	rolls  []singleThresholdRoll
}

func (t *ThresholdRoll) String() (string, error) {
	builder := strings.Builder{}
	for _, roll := range t.rolls {
		if roll.isExplode {
			_, err := builder.WriteString(fmt.Sprintf("(%d) ", roll.value))
			if err != nil {
				return "", err
			}
		} else {
			_, err := builder.WriteString(fmt.Sprintf("%d ", roll.value))
			if err != nil {
				return "", err
			}
		}
	}
	return strings.TrimSpace(builder.String()), nil
}

func (t *ThresholdRoll) Value() int {
	hits := 0
	for _, roll := range t.rolls {
		if roll.value >= t.params.passOn {
			hits++
		}
	}
	return hits
}

type ThresholdRoller struct {
	baseRoller *BaseRoller
}

func NewThresholdRoller() *ThresholdRoller {
	return &ThresholdRoller{
		NewBaseRoller(),
	}
}

func (t *ThresholdRoller) DoThresholdRoll(count, sides int, params ThresholdParameters) (ThresholdRoll, error) {
	result := ThresholdRoll{
		params: params,
		rolls:  make([]singleThresholdRoll, 0),
	}

	for i := 0; i < count; i++ {
		rolling := true
		wasExplode := false
		for rolling {
			roll := t.baseRoller.getRoll(sides)
			rolling = roll >= params.explodeOn // keep going if the roll explodes
			result.rolls = append(result.rolls, singleThresholdRoll{
				roll,
				wasExplode,
			})
			wasExplode = rolling
		}
	}

	return result, nil
}
