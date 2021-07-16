package roller

const (
	SrNoGlitch SrGlitchStatus = iota
	SrGlitch
	SrCritGlitch
)

const SrDieSides = 6

type SrGlitchStatus int

func GetSrParams() ThresholdParameters {
	return ThresholdParameters{
		passOn:    5,
		explodeOn: 6,
	}
}

func GetGlitchStatus(roll ThresholdRoll) SrGlitchStatus {
	if isGlitch(roll) {
		if roll.Value() == 0 {
			return SrCritGlitch
		} else {
			return SrGlitch
		}
	}
	return SrNoGlitch
}

func isGlitch(roll ThresholdRoll) bool {
	ones := 0
	for _, singleRoll := range roll.rolls {
		if singleRoll.value == 1 {
			ones++
		}
	}
	return ones > len(roll.rolls)/2
}
