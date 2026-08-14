package proxy

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func NewProxy(targetURL string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal("invalid service url")
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
