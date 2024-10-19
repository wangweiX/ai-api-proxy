package middleware

import (
	"ai-api-proxy/pkg/logger"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware verify api key
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

// SecurityHeadersMiddleware set security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing attacks.
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent click jacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		// Prevent cross-site scripting attacks
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		// Implement content security policy (CSP), limit resource loading.
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'none'")
		// Prevent browser caching
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "0")
		c.Next()
	}
}

// LimitRequestBody limit request body size
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

// ErrorHandler unified error handling function
func ErrorHandler(w http.ResponseWriter, errMsg string, statusCode int) {
	logger.Logger.Error(errMsg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": errMsg}
	_ = json.NewEncoder(w).Encode(response)
}
