package handler

import (
	"net/http"

	"github.com/Ghostbaby/sls-migrate/internal/service"
	"github.com/gin-gonic/gin"
)

// SLSHandler SLS 处理器
type SLSHandler struct {
	slsService  service.SLSService
	syncService service.SyncService
}

// NewSLSHandler 创建新的 SLSHandler 实例
func NewSLSHandler(slsService service.SLSService, syncService service.SyncService) *SLSHandler {
	return &SLSHandler{
		slsService:  slsService,
		syncService: syncService,
	}
}

// GetSLSAlerts 从阿里云 SLS 获取所有 Alert 规则
// @Summary 从阿里云 SLS 获取所有 Alert 规则
// @Description 从阿里云 SLS 获取所有 Alert 规则
// @Tags SLS
// @Accept json
// @Produce json
// @Success 200 {array} models.Alert
// @Failure 500 {object} map[string]interface{}
// @Router /sls/alerts [get]
func (h *SLSHandler) GetSLSAlerts(c *gin.Context) {
	alerts, err := h.slsService.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get alerts from SLS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  alerts,
		"count": len(alerts),
	})
}

// GetSLSAlertByName 根据名称从阿里云 SLS 获取特定 Alert 规则
// @Summary 根据名称从阿里云 SLS 获取特定 Alert 规则
// @Description 根据名称从阿里云 SLS 获取特定 Alert 规则
// @Tags SLS
// @Accept json
// @Produce json
// @Param name path string true "Alert 名称"
// @Success 200 {object} models.Alert
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /sls/alerts/name/{name} [get]
func (h *SLSHandler) GetSLSAlertByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert name",
			"message": "Name cannot be empty",
		})
		return
	}

	alert, err := h.slsService.GetAlertByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Alert not found in SLS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// SyncSLSAlerts 同步阿里云 SLS 的 Alert 规则到本地数据库
// @Summary 同步阿里云 SLS 的 Alert 规则到本地数据库
// @Description 同步阿里云 SLS 的 Alert 规则到本地数据库
// @Tags SLS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sls/sync [post]
func (h *SLSHandler) SyncSLSAlerts(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync service not available",
			"message": "Sync service is not initialized",
		})
		return
	}

	err := h.syncService.SyncSLSToDatabase(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync alerts from SLS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully synced alerts from SLS",
	})
}

// SyncDatabaseToSLS 同步本地数据库的 Alert 规则到阿里云 SLS
// @Summary 同步本地数据库的 Alert 规则到阿里云 SLS
// @Description 同步本地数据库的 Alert 规则到阿里云 SLS
// @Tags SLS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sls/sync/db-to-sls [post]
func (h *SLSHandler) SyncDatabaseToSLS(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync service not available",
			"message": "Sync service is not initialized",
		})
		return
	}

	err := h.syncService.SyncDatabaseToSLS(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to sync alerts to SLS",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully synced alerts to SLS",
	})
}

// GetSyncStatus 获取同步状态
// @Summary 获取同步状态
// @Description 获取同步状态
// @Tags SLS
// @Accept json
// @Produce json
// @Success 200 {object} service.SyncStatus
// @Failure 500 {object} map[string]interface{}
// @Router /sls/sync/status [get]
func (h *SLSHandler) GetSyncStatus(c *gin.Context) {
	if h.syncService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Sync service not available",
			"message": "Sync service is not initialized",
		})
		return
	}

	status, err := h.syncService.GetSyncStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get sync status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetSLSStatus 获取 SLS 连接状态
// @Summary 获取 SLS 连接状态
// @Description 获取 SLS 连接状态
// @Tags SLS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /sls/status [get]
func (h *SLSHandler) GetSLSStatus(c *gin.Context) {
	// 尝试获取一个 alert 来测试连接
	_, err := h.slsService.GetAlerts(c.Request.Context())

	status := "connected"
	message := "SLS connection is healthy"

	if err != nil {
		status = "disconnected"
		message = "SLS connection failed: " + err.Error()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"message": message,
	})
}
