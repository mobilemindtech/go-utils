package util

import (
	"time"
)

const (
	DateTimeDbLayout  = "2006-01-02T15:04:05-07:00"
	DateLayout        = "2006-01-02"
	TimeLayout        = "15:04:05"
	TimeMinutesLayout = "15:04"
	DateTimeLayout    = "2006-01-02 15:04:05"
	DateBrLayout      = "02/01/2006"
	DateTimeBrLayout  = "02/01/2006 15:04:05"
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

func ChangeDateWithTimeParts(date time.Time, hour int, minute int, second int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, second, date.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithTime(d1 time.Time, t1 time.Time) time.Time {
	return time.Date(d1.Year(), d1.Month(), d1.Day(), t1.Hour(), t1.Minute(), t1.Second(), d1.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithDate(d1 time.Time, d2 time.Time) time.Time {
	return time.Date(d2.Year(), d2.Month(), d2.Day(), d1.Hour(), d1.Minute(), d1.Second(), d1.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithDateParts(d1 time.Time, day int, month int, year int) time.Time {
	return time.Date(year, time.Month(month), day, d1.Hour(), d1.Minute(), d1.Second(), d1.Nanosecond(), GetDefaultLocation())
}


func ChangeDateWithHour(date time.Time, hour int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), hour, date.Minute(), date.Second(), date.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithMinute(date time.Time, minute int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), minute, date.Second(), date.Nanosecond(), GetDefaultLocation())
}

func ChangeDateWithSecond(date time.Time, second int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), second, date.Nanosecond(), GetDefaultLocation())
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

func AfterOrEqual(t1 time.Time , t2 time.Time) bool {
	return t1.After(t2) || IsSameDateTime(t1, t2)
}

func BeforeOrEqual(t1 time.Time , t2 time.Time) bool {
	return t1.Before(t2) || IsSameDateTime(t1, t2)
}

func Contains(begin time.Time, end time.Time, date time.Time) bool {
	return date.After(begin) && date.Before(end)
}
