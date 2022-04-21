package router

import (
	"github.com/gin-gonic/gin"
	"server/middleware"
	"server/router/api"
)

func Init() *gin.Engine {
	r := gin.New()
	// 允许跨域访问，有nginx等代理的可以关闭
	r.Use(middleware.Cors())
	// 设置可访问路由
	api.Extra(r)
	api.Depreciation(r)
	api.Block(r)
	api.Transaction(r)
	api.Exchanger(r)
	api.Collection(r)
	api.NFT(r)
	return r
}
