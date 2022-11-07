package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "server/docs"
	"server/middleware"
	"server/router/api"
)

func Run(addr string) error {
	r := gin.New()
	// Allow cross-domain access, and those with nginx and other proxies can be closed
	r.Use(middleware.Cors())
	// Set up accessible routes
	api.Extra(r)
	api.Depreciation(r)
	api.Block(r)
	api.Transaction(r)
	api.Account(r)
	api.Exchanger(r)
	api.Collection(r)
	api.NFT(r)
	api.Reward(r)
	api.Epoch(r)
	api.Ranking(r)
	api.Chart(r)
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(addr)
}
