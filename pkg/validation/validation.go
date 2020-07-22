package validation

import (
	"errors"
	"reflect"
	"regexp"
)

type Validation struct {
	status bool
	err    *MyError
}

func (v *Validation) setError(e *MyError) {
	v.err = e
	v.status = true
}

func (v *Validation) GetError() error {
	if v.err != nil {
		return errors.New(v.err.info)
	}

	return nil
}

func (v *Validation) GetErroString() string {
	if v.err != nil {
		return v.err.info
	}

	return ""
}

func (v *Validation) HasError() bool {
	var s bool
	if v.err != nil {
		s = true
	}

	return s
}

func (v *Validation) Required(obj interface{}) *Result {
	return v.apply(Required{}, obj)
}

func (v *Validation) Email(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^[\w-]+@[a-zA-Z0-9_-]{1,10}\.[\w]{2,8}$`))
}

func (v *Validation) Numeric(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^[0-9]+$`))
}

//只允许0或1
func (v *Validation) Switch(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^[0|1]$`))
}

//只允许数字或逗号
func (v *Validation) NumberOrComma(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^[0-9|,]+$`))
}

//手机号码
func (v *Validation) Mobile(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^1[0-9]{10}$`))
}

//大小写字母 数字
func (v *Validation) AlphaNumeric(obj interface{}) *Result {
	return v.Match(obj, regexp.MustCompile(`^[0-9a-zA-Z]+$`))
}

func (v *Validation) Match(obj interface{}, regex *regexp.Regexp) *Result {
	return v.apply(Match{regex}, obj)
}

func (v *Validation) apply(chk Validator, obj interface{}) *Result {
	res := &Result{}

	if v.status {
		return res
	}

	if nil == obj {
		if chk.verify(obj) {
			return res
		}
	} else if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		if reflect.ValueOf(obj).IsNil() {
			if chk.verify(nil) {
				return res
			}
		} else {
			if chk.verify(reflect.ValueOf(obj).Elem().Interface()) {
				return res
			}
		}
	} else if chk.verify(obj) {
		return res
	}

	e := &MyError{}
	v.setError(e)
	res.err = e

	return res
}
