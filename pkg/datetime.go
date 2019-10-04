package pkg

import "time"

const (
	IsoYearMonthDaySlash = "2006/01/02"
	IsoYearMonthDay      = "2006-01-02"
	IsoDateTime          = "2006-01-02T15:04:05.000-0700"
)

func GetTimeRange(year int, month time.Month) (time.Time, time.Time) {
	toYear := year
	toMonth := month + 1
	if month == time.December {
		toYear++
		toMonth = time.January
	}

	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC), time.Date(toYear, toMonth, 1, 0, 0, 0, 0, time.UTC)
}
