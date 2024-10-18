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

	"go.uber.org/zap"
)

// 定义需要保留和删除的头部
var keepHeaderLower = map[string]bool{
	"x-api-key":     true,
	"authorization": true,
}

// NewOpenAIReverseProxy 创建一个新的 OpenAI 代理
func NewOpenAIReverseProxy(cfg *config.Config) (*httputil.ReverseProxy, error) {
	director := func(req *http.Request) {
		// 获取请求的路径前缀
		var targetURL string
		for prefix, target := range cfg.PathMap {
			if strings.HasPrefix(req.URL.Path, prefix) {
				targetURL = target
				// 移除路径前缀
				req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
				break
			}
		}

		if targetURL == "" {
			logger.Logger.Warnf("未知的路径: %s", req.URL.Path)
			return
		}

		remote, err := url.Parse(targetURL)
		if err != nil {
			return
		}

		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.Host = remote.Host

		logger.Logger.Info("before request headers", zap.Any("headers", req.Header))

		// 保留必要的头部，删除其他不需要的头部
		for header := range req.Header {
			headerLower := strings.ToLower(header)
			if !keepHeaderLower[headerLower] {
				// 去除 Cf-Connecting-Ip、 Cf-Ipcountry 等cloudflare头信息
				if strings.HasPrefix(headerLower, "cf-") ||
					strings.EqualFold(headerLower, "cdn-loop") {
					req.Header.Del(header)
				}
			}
		}

		// 设置请求头部
		req.Header.Set("Host", remote.Host)

		if cfg.FixedRequestIP != "" {
			// 设置本机物理机IP，防止暴露原客户端IP
			req.Header.Set("X-Real-IP", cfg.FixedRequestIP)
			req.Header.Set("X-Forwarded-For", cfg.FixedRequestIP)
		}

		// 打印所有请求头部
		logger.Logger.Info("after request headers", zap.Any("headers", req.Header))
	}

	// 创建 HTTP 传输层，设置超时和连接池
	transport := &http.Transport{
		// 设置支持WebSocket,自动处理升级请求
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		// 处理WebSocket协议
		ModifyResponse: func(res *http.Response) error {
			if res.Header.Get("Upgrade") == "websocket" {
				res.Header.Del("Content-Length")
			}
			return nil
		},

		// 处理错误
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			middleware.ErrorHandler(w, fmt.Sprintf("代理请求失败: %v", err), http.StatusBadGateway)
		},
	}

	return proxy, nil
}
