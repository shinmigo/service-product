package myerror

const (
	//通用的错误提示
	CommonErrorOfErrorCode = -101

	//没有权限
	NoPermissionOfErrorCode = -120
	//没有找到
	NotFoundOfErrorCode = -404
	//服务器错误
	ServerInternalErrorOfErrorCode = -500
)

//错误码规则定义：一共6位，比如：110101， 第一第二位表示项目，第三第四位表示哪个controller, 最后二位表示这个控制器下面的错误码
//比如如下
var CodeList = map[int]string{
	-140101: "你好呀~~~~~~",
	-140102: "你好呀1",
	-140103: "你好呀2",
	-140104: "你好呀3",
	-140105: "你好呀4",
}
