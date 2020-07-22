package ctl

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
	AppController interface{}
	*gin.Context
}

func (gh *Controller) Init(ctl ControllerInterface, c *gin.Context) {
	gh.AppController = ctl
	gh.Context = c
}

func (gh *Controller) Prepare() {

}

func (gh *Controller) Finish() {

}
