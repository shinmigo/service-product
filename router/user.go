package router

import (
	"goshop/service-product/controller"
	"goshop/service-product/pkg/core/routerhelper"
	"goshop/service-product/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	routerhelper.Use(func(r *gin.Engine) {
		g := routerhelper.NewGroupRouter("user", new(controller.User), r, middleware.Cors(), middleware.Test())
		g.Get("/get-list-query", "GetListQuery")
	})
}
