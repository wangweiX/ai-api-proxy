package proxy

import (
	"ai-api-proxy/internal/config"
	"ai-api-proxy/internal/middleware"
	"ai-api-proxy/pkg/logger"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// keepHeaderLower keep some headers
var keepHeaderLower = map[string]bool{
	"x-api-key":     true,
	"authorization": true,
}

// NewOpenAIReverseProxy create a new OpenAI reverse proxy
func NewOpenAIReverseProxy(cfg *config.Config) (*httputil.ReverseProxy, error) {
	director := func(req *http.Request) {
		startTime := time.Now()
		logger.Logger.Infof("Start to handle request, method: %s, path: %s", req.Method, req.URL.Path)

		// Get the target URL based on the request path
		var targetURL string
		for prefix, target := range cfg.PathMap {
			if strings.HasPrefix(req.URL.Path, prefix) {
				targetURL = target
				// Remove the prefix from the request path
				req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
				break
			}
		}

		if targetURL == "" {
			logger.Logger.Warnf("Unknown path: %s", req.URL.Path)
			return
		}

		remote, err := url.Parse(targetURL)
		if err != nil {
			return
		}

		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.Host = remote.Host

		logger.Logger.Infof("Origin Request Headers, method: %s, path: %s, headers: %v", req.Method, req.URL.Path, req.Header)

		// Keep necessary headers, remove other unnecessary headers
		for header := range req.Header {
			headerLower := strings.ToLower(header)
			if !keepHeaderLower[headerLower] {
				// Remove Cf-Connecting-Ip、 Cf-Ipcountry 等cloudflare头信息
				if strings.HasPrefix(headerLower, "cf-") ||
					strings.EqualFold(headerLower, "cdn-loop") {
					req.Header.Del(header)
				}
			}
		}

		// Set the Host header
		req.Header.Set("Host", remote.Host)

		if cfg.FixedRequestIP != "" {
			// Set the physical machine IP to prevent exposure of the original client IP
			req.Header.Set("X-Real-IP", cfg.FixedRequestIP)
			req.Header.Set("X-Forwarded-For", cfg.FixedRequestIP)
		}

		// Print all request headers
		logger.Logger.Infof("Request processing completed, method: %s, path: %s, new request headers: %v, duration: %s", req.Method, req.URL.Path, req.Header, time.Since(startTime))
	}

	// Create HTTP transport layer, set timeout and connection pool
	transport := &http.Transport{
		// Set support for WebSocket, automatically handle upgrade requests
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     120 * time.Second,
		TLSHandshakeTimeout: 20 * time.Second,
	}

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		// Handle WebSocket protocol
		ModifyResponse: func(res *http.Response) error {
			if res.Header.Get("Upgrade") == "websocket" {
				res.Header.Del("Content-Length")
			}
			return nil
		},

		// Handle errors
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			middleware.ErrorHandler(w, fmt.Sprintf("Proxy request failed: %v", err), http.StatusBadGateway)
		},
	}

	return proxy, nil
}
