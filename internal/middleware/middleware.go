package middleware

import (
	"ai-api-proxy/pkg/logger"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware 验证请求中的 API 密钥
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")
		if apiKey == "" {
			apiKey = c.GetHeader("x-api-key")
		}
		if apiKey == "" {
			logger.Logger.Warn("未授权的访问请求")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware 设置安全头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 MIME 类型嗅探攻击。
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		// 防止点击劫持
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		// 防止跨站脚本攻击
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		// 实施内容安全策略（CSP），限制资源加载。
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'none'")
		// 防止浏览器缓存
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")
		c.Next()
	}
}

// LimitRequestBody 限制请求体大小
func LimitRequestBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		if err := c.Request.ParseForm(); err != nil {
			logger.Logger.Warn("请求体过大")
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "请求体过大"})
			return
		}
		c.Next()
	}
}

// ErrorHandler 统一的错误处理函数
func ErrorHandler(w http.ResponseWriter, errMsg string, statusCode int) {
	logger.Logger.Error(errMsg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": errMsg}
	_ = json.NewEncoder(w).Encode(response)
}
