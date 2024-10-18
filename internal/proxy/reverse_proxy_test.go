package proxy

import (
	"ai-api-proxy/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestNewOpenAIReverseProxy_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		FixedRequestIP: "203.0.113.1",
	}

	logger, _ := zap.NewDevelopment()
	defer func() {
		if err := logger.Sync(); err != nil {
			assert.Error(t, err)
		}
	}()

	proxy, err := NewOpenAIReverseProxy(cfg)
	assert.Error(t, err)
	assert.Nil(t, proxy)
}
