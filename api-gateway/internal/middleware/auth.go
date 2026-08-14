package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwt_secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "bearer token required"})
			ctx.Abort()
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "bearer token required"})
			ctx.Abort()
			return
		}
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwt_secret), nil
		})
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}
		role, ok := claims["role"].(string)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}
		ctx.Request.Header.Set("X-User-ID", userID)
		ctx.Request.Header.Set("X-User-Role", role)
		ctx.Next()
	}
}
