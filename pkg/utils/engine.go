package utils

import "github.com/gin-gonic/gin"

var R *gin.Engine

func NewGinDefault() *gin.Engine {
	r := gin.Default()
	R = r

	return r
}
