package util

import "time"

// GetLastMonth gets the string for the previous month in YYYY-MM format
func GetLastMonth(now time.Time) string {
	t := now.AddDate(0, -1, 0)
	return t.Format("2006-01")
}
