package contrib

import (
	"strings"
	"time"

	"github.com/azhai/gozzo-utils/calendar"
)

var (
	WeekEndType      = calendar.W_MON_FRI // 双休
	cal2019, cal2020 *calendar.Calendar
)

/**
 * 判断是否节假日，日期使用yyyy-mm-dd格式
 */
func IsHoliday(date string) bool {
	if strings.HasPrefix(date, "2019") {
		if cal2019 == nil {
			cal2019 = calendar.NewYearCalendar(2019, WeekEndType)
			cal2019 = SetCalendarY2019(cal2019)
		}
		return cal2019.IsHoliday(date)
	} else if strings.HasPrefix(date, "2020") {
		if cal2020 == nil {
			cal2020 = calendar.NewYearCalendar(2020, WeekEndType)
			cal2020 = SetCalendarY2020(cal2020)
		}
		return cal2020.IsHoliday(date)
	}
	panic("The year not in our calendars !")
}

/**
 * 找出下一个工作日（不含今天）
 */
func GetNextWorkday(t time.Time) (time.Time, bool) {
	for {
		t.AddDate(0, 0, 1)
		year := t.Year()
		if year > 2020 || year < 2019 {
			break
		}
		if IsHoliday(t.Format("2006-01-02")) {
			continue
		}
		return t, true
	}
	return time.Time{}, false
}

/**
 * 2019年节日调休年历
 */
func SetCalendarY2019(cal *calendar.Calendar) *calendar.Calendar {
	// 元旦
	cal.SetHoliday("2019-01-01")
	// 春节
	cal.SetWorkday("2019-02-02")
	cal.SetWorkday("2019-02-03")
	cal.SetHoliday("2019-02-04")
	cal.SetHoliday("2019-02-05")
	cal.SetHoliday("2019-02-06")
	cal.SetHoliday("2019-02-07")
	cal.SetHoliday("2019-02-08")
	// 清明节
	cal.SetHoliday("2019-04-05")
	// 劳动节
	cal.SetWorkday("2019-04-28")
	cal.SetHoliday("2019-05-01")
	cal.SetHoliday("2019-05-02")
	cal.SetHoliday("2019-05-03")
	cal.SetWorkday("2019-05-05")
	// 端午节
	cal.SetHoliday("2019-06-07")
	// 中秋节
	cal.SetHoliday("2019-09-13")
	// 国庆节
	cal.SetWorkday("2019-09-29")
	cal.SetHoliday("2019-10-01")
	cal.SetHoliday("2019-10-02")
	cal.SetHoliday("2019-10-03")
	cal.SetHoliday("2019-10-04")
	cal.SetHoliday("2019-10-07")
	cal.SetWorkday("2019-10-12")
	return cal
}

/**
 * 2020年节日调休年历
 */
func SetCalendarY2020(cal *calendar.Calendar) *calendar.Calendar {
	// 元旦
	cal.SetHoliday("2020-01-01")
	// 春节
	cal.SetWorkday("2020-01-19")
	cal.SetHoliday("2020-01-24")
	cal.SetHoliday("2020-01-25")
	cal.SetHoliday("2020-01-27")
	cal.SetHoliday("2020-01-28")
	cal.SetHoliday("2020-01-29")
	cal.SetHoliday("2020-01-30")
	cal.SetWorkday("2020-02-01")
	// 清明节
	cal.SetHoliday("2020-04-04")
	cal.SetHoliday("2020-04-06")
	// 劳动节
	cal.SetWorkday("2020-04-26")
	cal.SetHoliday("2020-05-01")
	cal.SetHoliday("2020-05-02")
	cal.SetHoliday("2020-05-04")
	cal.SetHoliday("2020-05-05")
	cal.SetWorkday("2020-05-09")
	// 端午节
	cal.SetHoliday("2020-06-25")
	cal.SetHoliday("2020-06-26")
	cal.SetHoliday("2020-06-27")
	cal.SetWorkday("2020-06-28")
	// 中秋节、国庆节
	cal.SetHoliday("2020-10-01")
	cal.SetHoliday("2020-10-02")
	cal.SetHoliday("2020-10-03")
	cal.SetHoliday("2020-10-05")
	cal.SetHoliday("2020-10-06")
	cal.SetHoliday("2020-10-07")
	cal.SetHoliday("2020-10-08")
	cal.SetWorkday("2020-10-10")
	return cal
}
