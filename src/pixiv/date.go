package pixiv

import (
	"fmt"
	"time"
)

func WeekByDate(t time.Time) string {
	yearDay := t.YearDay()
	yearFirstDay := t.AddDate(0, 0, -yearDay+1)
	firstDayInWeek := int(yearFirstDay.Weekday())

	//今年第一周有几天
	firstWeekDays := 1
	if firstDayInWeek != 0 {
		firstWeekDays = 7 - firstDayInWeek + 1
	}
	var week int
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		week = (yearDay-firstWeekDays)/7 + 2
	}
	return string(fmt.Sprintf("%d", week))
}

func DateFormat(mode string) string {
	if mode == "weekly" {
		return fmt.Sprintf("-%d-%s", time.Now().Year(), WeekByDate(time.Now()))
	} else if mode == "monthly" {
		return time.Now().Format("-2006-01")
	} else {
		return time.Now().Format("-2006-01-02")
	}
}
