package api

import (
	"github.com/gin-gonic/gin"
)

type User struct {
	*gin.Context
}

func NewUser(c *gin.Context) *User {
	return &User{Context: c}
}

func (m *User) GetListQuery(p string) (string, error) {
	//相关业务逻辑
	name := "hello" + p

	return name, nil
}
