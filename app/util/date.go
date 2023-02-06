package util

import (
	"time"
)

func GetDefaultLocation() *time.Location {
	location, _ := time.LoadLocation("America/Sao_Paulo")
	return location
}

func DateParse(layout string, data string) (time.Time, error) {
	return time.ParseInLocation(layout, data, GetDefaultLocation())
}

func DateNow() time.Time {
	return time.Now().In(GetDefaultLocation())
}

func IsAfterToday(d time.Time) bool {
	now := ClearTime(DateNow())
	d = ClearTime(d)
	return d.After(now)
}

func IsBeforeToday(d time.Time) bool {
	now := ClearTime(DateNow())
	d = ClearTime(d)
	return d.Before(now)
}

func DateNowZeroTime() time.Time {
	return ClearTime(time.Now().In(GetDefaultLocation()))
}

func DateFormat(date time.Time, layout string) time.Time {
	return date.In(GetDefaultLocation())
}

func GetTodayWithTime(newTime time.Time) time.Time {
	now := DateNow()
	return time.Date(now.Year(), now.Month(), now.Day(), newTime.Hour(), newTime.Minute(), newTime.Second(), newTime.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithDay(date time.Time, day int) time.Time {
	return time.Date(date.Year(), date.Month(), day, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithMonth(date time.Time, month int) time.Time {
	return time.Date(date.Year(), time.Month(month), date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithYear(date time.Time, year int) time.Time {
	return time.Date(year, date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), GetDefaultLocation())
}

func GetTodayWithStrTime(layout string, newTime string) (time.Time, error) {
	d, err := DateParse(layout, newTime)

	if err != nil {
		return time.Now(), err
	}

	return GetTodayWithTime(d), nil
}

func BeginningOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, GetDefaultLocation())
}

func EndOfMonth(date time.Time) time.Time {
	//.AddDate(0, 1, -1)
	return BeginningOfMonth(date).AddDate(0, 1, 0).Add(time.Nanosecond * -1)
}

func ClearTimeNow() time.Time {
	return ClearTime(DateNow())
}

func ClearTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, GetDefaultLocation())
}

func FirstTimeNow() time.Time {
	return FirstTime(DateNow())
}

func FirstTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, GetDefaultLocation())
}

func LastTimeNow() time.Time {
	return LastTime(DateNow())
}

func LastTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, GetDefaultLocation())
}

func IsSameDate(date1 time.Time, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month() && date1.Day() == date2.Day()
}

func IsSameTime(date1 time.Time, date2 time.Time) bool {
	return date1.Hour() == date2.Hour() && date1.Minute() == date2.Minute()
}

func IsSameMonth(date1 time.Time, date2 time.Time) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month()
}

func IsSameDateTime(date1 time.Time, date2 time.Time) bool {
	return IsSameDate(date1, date2) && IsSameTime(date1, date2)
}

func ContainsWithSame(begin time.Time, end time.Time, date time.Time) bool {
	return IsSameDateTime(begin, date) || IsSameDateTime(end, date) || (date.After(begin) && date.Before(end))
}

func Contains(begin time.Time, end time.Time, date time.Time) bool {
	return date.After(begin) && date.Before(end)
}
