package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Ghostbaby/sls-migrate/internal/config"
	"github.com/Ghostbaby/sls-migrate/internal/handler"
	"github.com/Ghostbaby/sls-migrate/internal/service"
	"github.com/Ghostbaby/sls-migrate/internal/store"
	"github.com/Ghostbaby/sls-migrate/pkg/database"
)

// @title SLS Migrate API
// @version 1.0
// @description SLS Alert 迁移系统的 API 接口
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @schemes http https
func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	if err := database.InitDatabase(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// 创建依赖
	alertStore := store.NewAlertStore()
	alertService := service.NewAlertService(alertStore)
	alertHandler := handler.NewAlertHandler(alertService)

	// 创建 SLS 服务
	slsConfig := config.LoadSLSConfig()
	slsService, err := service.NewSLSService(slsConfig)
	if err != nil {
		log.Printf("Warning: Failed to create SLS service: %v", err)
		log.Println("SLS functionality will be disabled")
		slsService = nil
	}

	// 创建同步服务
	var syncService service.SyncService
	if slsService != nil {
		syncService = service.NewSyncService(slsService, alertStore, alertService)
	}

	// 创建 SLS 处理器
	var slsHandler *handler.SLSHandler
	if slsService != nil {
		slsHandler = handler.NewSLSHandler(slsService, syncService)
	} else {
		// 创建一个空的处理器，避免 panic
		slsHandler = &handler.SLSHandler{}
	}

	// 设置路由
	router := handler.SetupRouter(alertHandler, slsHandler)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
