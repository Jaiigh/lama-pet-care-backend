package utils

import (
	"fmt"
	"time"
)

func GetTimeBangKokZone() time.Time {
	// Set the Bangkok, Thailand time zone
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return time.Now().Add(7 * time.Hour)
	}

	return time.Now().In(loc)
}
