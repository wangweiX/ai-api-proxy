package proxy

import (
	"ai-api-proxy/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOpenAIReverseProxy_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		FixedRequestIP: "203.0.113.1",
	}

	proxy, err := NewOpenAIReverseProxy(cfg)
	assert.Error(t, err)
	assert.Nil(t, proxy)
}
