package routers

import (
	"github.com/gin-gonic/gin"
	"server/handlers/block"
	"server/middleware"
)

func init() {
	include(block.Routers)
}

type Option func(*gin.Engine)

var options []Option

func include(opts ...Option) {
	options = append(options, opts...)
}

func Init() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Cors())
	for _, opt := range options {
		opt(r)
	}
	return r
}
