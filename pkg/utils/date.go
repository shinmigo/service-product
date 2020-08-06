package utils

import (
	"strconv"
	"time"
)

//获取每月的几号
func GetMonthDay() string {
	localZone, _ := time.LoadLocation("Asia/Shanghai")
	dayInt := time.Now().In(localZone).Day()
	return strconv.Itoa(dayInt)
}

/*
获取星期几
注意：如果是周日，系统返回的是0，mms系统的规则是：1,2,3,4,5,6,7
即把0重新赋为7
*/
func GetWeekday() string {
	localZone, _ := time.LoadLocation("Asia/Shanghai")
	buf := int(time.Now().In(localZone).Weekday())
	wday := strconv.Itoa(buf)
	if wday == "0" {
		wday = "7"
	}

	return wday
}
