package routerhelper

import (
	"goshop/service-product/pkg/utils"

	"github.com/gin-gonic/gin"
)

type RouterFun func(r *gin.Engine)

var rList = make([]RouterFun, 0, 8)

func Use(p ...RouterFun) {
	rList = append(rList, p...)
}

func EntryRouterTree(e *gin.Engine) {
	for k := range rList {
		rList[k](e)
	}
}

func InitRouter() {
	r := utils.NewGinDefault()
	EntryRouterTree(r)
}
