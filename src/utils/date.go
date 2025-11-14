package utils

import "time"

func CheckSameDate(LHSDate, RHSDate time.Time) bool {
	return LHSDate.Year() == RHSDate.Year() &&
		LHSDate.Month() == RHSDate.Month() &&
		LHSDate.Day() == RHSDate.Day()
}

func GetRDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	startDate = startDate.UTC()
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second).UTC()
	return startDate, endDate, nil
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
