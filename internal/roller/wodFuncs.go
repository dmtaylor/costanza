package roller

import "strings"

const WodDieSides = 10

func GetWodRollParams(input []string) (params ThresholdParameters, isChance bool, rollStr string, err error) {
	err = nil
	rollStr = ""
	isChance = false
	params = ThresholdParameters{
		passOn:    8,
		explodeOn: 10,
	}
	builder := strings.Builder{}
	for _, item := range input {
		switch item {
		case "8again":
			params.explodeOn = 8
		case "9again":
			params.explodeOn = 9
		case "chance":
			isChance = true
		default:
			_, err = builder.WriteString(item + " ")
			if err != nil {
				return
			}
		}
	}
	rollStr = strings.TrimSpace(builder.String())
	return
}
