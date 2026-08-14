package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func RestaurantProxy() gin.HandlerFunc {
	target, _ := url.Parse("http://localhost:8081")
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
