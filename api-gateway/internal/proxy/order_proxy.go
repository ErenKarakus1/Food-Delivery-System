package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func OrderProxy() gin.HandlerFunc {
	target, _ := url.Parse("http://localhost:8082")
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
