package utils

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

var RequestRunStartTime time.Time

//时间格式
const TIME_STD_FORMART = "2006-01-02 15:04:05"
const TIME_STD_DATE_FORMART = "2006-01-02"
const TIME_STD_NO_FORMART = "20060102150405"

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

func GetNow() time.Time {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(location)
}

func (t *JSONTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		//t.Time = time.Time{}
		return
	}

	//t.Time, err = time.Parse(TIME_STD_FORMART, s)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t.Time, err = time.ParseInLocation(TIME_STD_FORMART, s, loc)

	return
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t JSONTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format(TIME_STD_FORMART))
	return []byte(formatted), nil
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
