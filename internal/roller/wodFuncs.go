package roller

const WodDieSides = 10

func NewGetWodRollParams(isNineAgain, isEightAgain bool) ThresholdParameters {
	params := ThresholdParameters{
		passOn:    8,
		explodeOn: 10,
	}
	if isNineAgain {
		params.explodeOn = 9
	}
	if isEightAgain {
		params.explodeOn = 8
	}
	return params
}
