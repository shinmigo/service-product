package controller

import (
	"fmt"
	"goshop/service-product/pkg/core/ctl"
	"goshop/service-product/pkg/myerror"
	"goshop/service-product/pkg/utils"
	"time"
)

type Base struct {
	RunStartTime time.Time
	ctl.Controller
}

//子类初始化
type InitialiseInterface interface {
	Initialise()
}

func (m *Base) Prepare() {
	m.RunStartTime = time.Now()
	if app, ok := m.AppController.(InitialiseInterface); ok {
		app.Initialise()
	}
	//这里可以做一些初始化的工作, 比如权限的验证，签名验证等等
}

func (m *Base) Finish() {
	fmt.Println(5, "Finish")
	//在这个方法可以，可以做一些收尾的工作，比如记录访问接口的日志
}

/**
params1 data
params2 message
params3 code   负数类(即小于0的)的表示失败，1表示成功.
code可以在调用setResponse时设置，这个是全局设置，如：
m.SetResponse(nil, "error message", -110122)

也可以在每个错误信息时，设置code,表示局部的错误信息, 如：
a := utils.NewErrorCode(-110123, "error message")
m.SetResponse(nil, a)

或者重置局部code
a := utils.NewErrorCode(-110123, "error message")
m.SetResponse(nil, a, -110122) 这样的话会打印-110122, 不会打印-110123

params4 other  其他的一些参数，用于方法中要记的日志
*/
func (m *Base) SetResponse(params ...interface{}) {
	var code = 1
	var message string
	var data interface{}

	lenBuf := len(params)
	//如果没有传参数，data就设置空
	if lenBuf == 0 {
		data = struct{}{}
	}

	//如果有一个，就把第一个设置到data中
	if lenBuf > 0 {
		if params[0] != nil {
			data = params[0]
		} else {
			data = struct{}{}
		}
	} else {
		data = struct{}{}
	}

	if lenBuf > 1 {
		isCommomError := true
		switch params[1].(type) {
		case string:
			message = params[1].(string)
		case error:
			message = params[1].(error).Error()
		case *utils.ErrorCode:
			b := params[1].(*utils.ErrorCode)
			if len(b.Content) == 0 {
				if v, ok := myerror.CodeList[b.CodeId]; ok {
					message = v
				}
			} else {
				message = b.Content
			}

			code = b.CodeId
			isCommomError = false

		default:
			message = "conversion is error"
		}

		if isCommomError {
			code = myerror.CommonErrorOfErrorCode
		}
	}

	if lenBuf > 2 {
		code = params[2].(int)
	}

	responseList := utils.ResponseList{
		RunTime: time.Since(m.RunStartTime).Seconds(),
		Code:    code,
		Message: message,
		Data:    data,
	}

	m.JSON(200, responseList)
}
