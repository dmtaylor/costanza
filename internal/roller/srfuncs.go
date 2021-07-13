package roller

func GetSrParams() ThresholdParameters {
	return ThresholdParameters{
		passOn:    5,
		explodeOn: 6,
	}
}

// TODO helper function for critical failures
func IsSRCritFail(roll ThresholdRoll) bool {
	//TODO do this
	return false
}
