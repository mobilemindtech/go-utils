
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

func GetTodayWithTime(newTime time.Time) time.Time{ 
	now := DateNow()
	return time.Date(now.Year(), now.Month(), now.Day(), newTime.Hour(), newTime.Minute(), newTime.Second(), newTime.Nanosecond(), GetDefaultLocation())
}

func GetTodayWithStrTime(layout string, newTime string) (time.Time, error){ 
	d, err := DateParse(layout, newTime)  

	if err != nil {
		return time.Now(), err
	}

	return GetTodayWithTime(d), nil
}

func BeginningOfMonth(date time.Time) time.Time{ 	
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, GetDefaultLocation())
}

func EndOfMonth(date time.Time) time.Time{ 
	//.AddDate(0, 1, -1)
	return BeginningOfMonth(date).AddDate(0, 1, 0).Add(time.Nanosecond * -1)	
}

  