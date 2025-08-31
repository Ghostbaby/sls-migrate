package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(alertHandler *AlertHandler, slsHandler *SLSHandler) *gin.Engine {
	router := gin.Default()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// API 路由组
	api := router.Group("/api/v1")
	{
		// Alert 相关路由
		alerts := api.Group("/alerts")
		{
			alerts.POST("", alertHandler.CreateAlert)                      // 创建 Alert
			alerts.GET("", alertHandler.ListAlerts)                        // 获取 Alert 列表
			alerts.GET("/:id", alertHandler.GetAlertByID)                  // 根据 ID 获取 Alert
			alerts.GET("/name/:name", alertHandler.GetAlertByName)         // 根据名称获取 Alert
			alerts.PUT("/:id", alertHandler.UpdateAlert)                   // 更新 Alert
			alerts.DELETE("/:id", alertHandler.DeleteAlert)                // 删除 Alert
			alerts.GET("/status/:status", alertHandler.ListAlertsByStatus) // 根据状态获取 Alert 列表
		}

		// SLS 相关路由
		sls := api.Group("/sls")
		{
			sls.GET("/alerts", slsHandler.GetSLSAlerts)                 // 从 SLS 获取所有 Alert
			sls.GET("/alerts/name/:name", slsHandler.GetSLSAlertByName) // 从 SLS 根据名称获取 Alert
			sls.POST("/sync", slsHandler.SyncSLSAlerts)                 // 同步 SLS Alert 到数据库
			sls.POST("/sync/db-to-sls", slsHandler.SyncDatabaseToSLS)   // 同步数据库 Alert 到 SLS
			sls.GET("/sync/status", slsHandler.GetSyncStatus)           // 获取同步状态
			sls.GET("/status", slsHandler.GetSLSStatus)                 // 获取 SLS 连接状态
		}
	}

	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "SLS Migrate Service is running",
		})
	})

	return router
}
