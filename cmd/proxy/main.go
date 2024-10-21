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
	configPath := flag.String("config", "config.yaml", "config file path")
	serverPort := flag.Int("port", 3002, "server port")

	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	// Initialize logger
	err = logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Init logger failed: %v", err)
	}

	// Set up rate limiting
	rate, err := limiter.NewRateFromFormatted(cfg.RateLimit)
	if err != nil {
		logger.Logger.Fatalf("Rate limiting config error: %v", err)
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	limiterMiddleware := ginmiddleware.NewMiddleware(instance)

	// Set GIN to Release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	router := gin.New()
	// Add health check route
	router.GET("/generate_204", func(c *gin.Context) {
		logger.Logger.Infoln("generate_204 success.")
		c.Status(http.StatusNoContent)
	})

	// Initialize proxy
	openAIProxy, err := proxy.NewOpenAIReverseProxy(cfg)
	if err != nil {
		logger.Logger.Fatalf("Initialize proxy failed: %v", err)
	}

	// Apply rate limiting middleware
	apiGroup := router.Group("/")
	apiGroup.Use(middleware.APIKeyAuthMiddleware())
	apiGroup.Use(limiterMiddleware)
	apiGroup.Use(middleware.SecurityHeadersMiddleware())
	apiGroup.Use(middleware.LimitRequestBody(int64(cfg.MaxRequestBodySizeMB << 20)))

	for prefix := range cfg.PathMap {
		apiGroup.Any(prefix+"/*path", func(c *gin.Context) {
			openAIProxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// Set NoRoute handler on main router
	router.NoRoute(func(c *gin.Context) {
		middleware.ErrorHandler(c.Writer, fmt.Sprintf("Unknown path: %s", c.Request.URL.Path), http.StatusNotFound)
	})

	// Set up server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *serverPort),
		Handler:      router,
		ReadTimeout:  90 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errChan := make(chan error, 1)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatalf("Failed to start server: %v", err)
			errChan <- err
		} else {
			logger.Logger.Infof("Server started successfully: %d", *serverPort)
		}
	}()

	select {
	case <-quit:
		logger.Logger.Info("Received shutdown signal, shutting down server...")
	case err := <-errChan:
		logger.Logger.Errorf("Server error: %v", err)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatalf("Server shutdown failed: %v", err)
	}
	logger.Logger.Info("Server shutdown successfully")
}
