package util

import "strconv"

func IntSliceToStr(in []int) []string {
	result := make([]string, len(in))
	for i, value := range in {
		result[i] = strconv.Itoa(value)
	}
	return result
}

func StrMap(in []string, f func(string) string) []string {
	result := make([]string, len(in))
	for i, value := range in {
		result[i] = f(value)
	}
	return result
}
