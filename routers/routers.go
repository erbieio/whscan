package routers

import (
	"github.com/gin-gonic/gin"
	"server/handlers/NFT"
	"server/handlers/SNFT"
	"server/handlers/block"
	"server/handlers/exchanger"
	"server/handlers/extra"
	"server/middleware"
)

func init() {
	include(block.Routers)
	include(extra.Routers)
	include(exchanger.Routers)
	include(NFT.Routers)
	include(SNFT.Routers)
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
