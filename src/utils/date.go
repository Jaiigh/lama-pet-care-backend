package utils

import "time"

func CheckSameDate(LHSDate, RHSDate time.Time) bool {
	return LHSDate.Year() == RHSDate.Year() &&
		LHSDate.Month() == RHSDate.Month() &&
		LHSDate.Day() == RHSDate.Day()
}
