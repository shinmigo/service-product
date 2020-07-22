package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Validator interface {
	verify(interface{}) bool
}

type MyError struct {
	info string
}

type Result struct {
	err *MyError
}

func (r *Result) Message(err string) {
	if r.err != nil {
		r.err.info = err
	}
}

type Required struct {
}

func (r Required) verify(obj interface{}) bool {
	if obj == nil {
		return false
	}

	if str, ok := obj.(string); ok {
		return len(strings.TrimSpace(str)) > 0
	}
	if _, ok := obj.(bool); ok {
		return true
	}
	if i, ok := obj.(int); ok {
		return i != 0
	}
	if i, ok := obj.(uint); ok {
		return i != 0
	}
	if i, ok := obj.(int8); ok {
		return i != 0
	}
	if i, ok := obj.(uint8); ok {
		return i != 0
	}
	if i, ok := obj.(int16); ok {
		return i != 0
	}
	if i, ok := obj.(uint16); ok {
		return i != 0
	}
	if i, ok := obj.(uint32); ok {
		return i != 0
	}
	if i, ok := obj.(int32); ok {
		return i != 0
	}
	if i, ok := obj.(int64); ok {
		return i != 0
	}
	if i, ok := obj.(uint64); ok {
		return i != 0
	}
	if t, ok := obj.(time.Time); ok {
		return !t.IsZero()
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		return v.Len() > 0
	}

	return true
}

type Match struct {
	Regexp *regexp.Regexp
}

func (m Match) verify(obj interface{}) bool {
	return m.Regexp.MatchString(fmt.Sprintf("%v", obj))
}
