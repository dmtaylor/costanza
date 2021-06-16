package util

import "strconv"

func IntSliceToStr(in []int) []string {
	result := make([]string, len(in))
	for i, value := range in {
		result[i] = strconv.Itoa(value)
	}
	return result
}
