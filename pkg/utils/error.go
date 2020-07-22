package utils

import "bytes"

type ErrorCode struct {
	CodeId  int
	Content string
}

func NewErrorCode(codeId int, content ...string) *ErrorCode {
	buf := bytes.NewBuffer(nil)
	for _, v := range content {
		buf.WriteString(v)
	}

	return &ErrorCode{
		CodeId:  codeId,
		Content: buf.String(),
	}
}
