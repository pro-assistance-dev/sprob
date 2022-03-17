package helper

import (
	"time"
)

func (h *Helper) GetMonthDays() []time.Time {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	days := make([]time.Time, 0)
	for {
		firstOfMonth = firstOfMonth.AddDate(0, 0, 1)
		days = append(days, firstOfMonth)
		if firstOfMonth.After(lastOfMonth) {
			break
		}
	}
	return days
}
