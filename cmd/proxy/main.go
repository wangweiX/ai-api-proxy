package main

import (
	"ai-api-proxy/internal/config"
	"ai-api-proxy/internal/middleware"
	"ai-api-proxy/internal/proxy"
	"ai-api-proxy/pkg/logger"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginmiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger.InitLogger(cfg)

	// 设置限流
	rate, err := limiter.NewRateFromFormatted(cfg.RateLimit)
	if err != nil {
		logger.Logger.Fatalf("限流配置错误: %v", err)
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	limiterMiddleware := ginmiddleware.NewMiddleware(instance)

	// 设置 GIN 为 Release 模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化 Gin 路由
	router := gin.New()
	// 添加健康检查路由
	router.GET("/generate_204", func(c *gin.Context) {
		logger.Logger.Debugln("generate_204 success.")
		c.Status(http.StatusNoContent)
	})

	// 初始化代理
	openAIProxy, err := proxy.NewOpenAIReverseProxy(cfg)
	if err != nil {
		logger.Logger.Fatalf("初始化代理失败: %v", err)
	}

	// 应用限流中间件
	apiGroup := router.Group("/")
	apiGroup.Use(middleware.APIKeyAuthMiddleware())
	apiGroup.Use(limiterMiddleware)
	apiGroup.Use(middleware.SecurityHeadersMiddleware())
	apiGroup.Use(middleware.LimitRequestBody(int64(cfg.MaxRequestBodySizeMB << 20)))
	//router.Use(middleware.ContentTypeMiddleware(logger.Logger))

	for prefix := range cfg.PathMap {
		apiGroup.Any(prefix+"/*path", func(c *gin.Context) {
			openAIProxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// 在主路由器上设置 NoRoute 处理程序
	router.NoRoute(func(c *gin.Context) {
		middleware.ErrorHandler(c.Writer, fmt.Sprintf("未知的路径: %s", c.Request.URL.Path), http.StatusNotFound)
	})

	// 设置服务器
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errChan := make(chan error, 1)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatalf("无法启动服务器: %v", err)
			errChan <- err
		} else {
			logger.Logger.Infof("服务启动成功：%s", cfg.ServerPort)
		}
	}()

	select {
	case <-quit:
		logger.Logger.Info("收到关闭信号，正在关闭服务器...")
	case err := <-errChan:
		logger.Logger.Errorf("服务器运行时发生错误: %v", err)
		quit <- syscall.SIGINT
	}

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatalf("服务器关闭失败: %v", err)
	}
	logger.Logger.Info("服务器已关闭")
}
